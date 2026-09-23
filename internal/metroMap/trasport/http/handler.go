package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/application"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
)

type Handler struct {
	ListPoints *application.ListPoint
}

func NewHandler(
	ListPoints *application.ListPoint,
) *Handler {
	return &Handler{
		ListPoints: ListPoints,
	}
}

func (h *Handler) Routes() chi.Router {

	r := chi.NewRouter()
	r.Get("/", h.getDatosMapa)

	return r
}

func (h *Handler) getDatosMapa(
	w http.ResponseWriter,
	r *http.Request,
) {

	result, err := h.ListPoints.Execute(
		r.Context(),
	)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	httpx.WriteJSON(
		w, http.StatusOK, result,
	)
}

func (h *Handler) writeDomainError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		domain.ErrorPunto,
	):
		httpx.WriteError(
			w, http.StatusFound,
			"punto no funciona",
		)

	default:
		httpx.WriteError(
			w,
			http.StatusInternalServerError,
			"internla server error",
		)
	}
}
