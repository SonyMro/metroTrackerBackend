package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type Handler struct {
	createUser     *application.CreateUser
	getUser        *application.GetUser
	listUsers      *application.ListUsers
	updateUser     *application.UpdateUser
	deactivateUser *application.DeactivateUser
}

func NewHandler(
	createUser *application.CreateUser,
	getUser *application.GetUser,
	listUsers *application.ListUsers,
	updateUser *application.UpdateUser,
	deactivateUser *application.DeactivateUser,
) *Handler {
	return &Handler{
		createUser:     createUser,
		getUser:        getUser,
		listUsers:      listUsers,
		updateUser:     updateUser,
		deactivateUser: deactivateUser,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.create)

	r.Get("/", h.list)

	r.Get("/{id}", h.get)

	r.Put("/{id}", h.update)

	r.Delete("/{id}", h.deactivate)

	return r
}

func (h *Handler) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request CreateUserRequest

	if err := decodeJSON(
		r,
		&request,
	); err != nil {
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)

		return
	}

	user := &domain.User{
		Nombre:   request.Nombre,
		Apepat:   request.Apepat,
		Apemat:   request.Apemat,
		Correo:   request.Correo,
		Password: request.Password,
		IDRol:    request.IDRol,
		IDGenero: request.IDGenero,
	}

	created, err := h.createUser.Execute(
		r.Context(),
		user,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	httpx.WriteJSON(
		w,
		http.StatusCreated,
		toUserResponse(created),
	)
}

func (h *Handler) get(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := pathID(r)

	if err != nil {
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)

		return
	}

	user, err := h.getUser.Execute(
		r.Context(),
		id,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	httpx.WriteJSON(
		w,
		http.StatusOK,
		toUserResponse(user),
	)
}

func (h *Handler) list(
	w http.ResponseWriter,
	r *http.Request,
) {
	page := queryInt(
		r,
		"page",
		1,
	)

	pageSize := queryInt(
		r,
		"pageSize",
		20,
	)

	result, err := h.listUsers.Execute(
		r.Context(),
		page,
		pageSize,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	items := make(
		[]*UserResponse,
		0,
		len(result.Items),
	)

	for _, user := range result.Items {
		items = append(
			items,
			toUserResponse(user),
		)
	}

	response := UserListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}

	httpx.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (h *Handler) update(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := pathID(r)

	if err != nil {
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)

		return
	}

	var request UpdateUserRequest

	if err := decodeJSON(
		r,
		&request,
	); err != nil {
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)

		return
	}

	err = h.updateUser.Execute(
		r.Context(),
		id,
		request.Nombre,
		request.Apepat,
		request.Apemat,
		request.IDRol,
		request.IDGenero,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	user, err := h.getUser.Execute(
		r.Context(),
		id,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	httpx.WriteJSON(
		w,
		http.StatusOK,
		toUserResponse(user),
	)
}

func (h *Handler) deactivate(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := pathID(r)

	if err != nil {
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)

		return
	}

	err = h.deactivateUser.Execute(
		r.Context(),
		id,
	)

	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	w.WriteHeader(
		http.StatusNoContent,
	)
}

func (h *Handler) writeDomainError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		domain.ErrUserNotFound,
	):
		httpx.WriteError(
			w,
			http.StatusNotFound,
			"user not found",
		)

	case errors.Is(
		err,
		domain.ErrEmailAlreadyExists,
	):
		httpx.WriteError(
			w,
			http.StatusConflict,
			"email already exists",
		)

	case errors.Is(
		err,
		domain.ErrInvalidUser,
	):
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid user",
		)

	case errors.Is(
		err,
		domain.ErrInvalidPassword,
	):
		httpx.WriteError(
			w,
			http.StatusBadRequest,
			"invalid password",
		)

	default:
		httpx.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func decodeJSON(
	r *http.Request,
	target any,
) error {
	decoder := json.NewDecoder(
		r.Body,
	)

	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func pathID(
	r *http.Request,
) (int32, error) {
	rawID := chi.URLParam(
		r,
		"id",
	)

	value, err := strconv.ParseInt(
		rawID,
		10,
		32,
	)

	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

func queryInt(
	r *http.Request,
	name string,
	fallback int,
) int {
	raw := r.URL.Query().Get(name)

	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)

	if err != nil {
		return fallback
	}

	return value
}
