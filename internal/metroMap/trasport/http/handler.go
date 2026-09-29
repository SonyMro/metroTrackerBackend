package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/application"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
)

// Handler expone los casos de uso del mapa mediante endpoints HTTP.
type Handler struct {
	ListPoints *application.ListPoint
}

// NewHandler crea el handler del mapa con el caso de uso para listar puntos.
func NewHandler(
	ListPoints *application.ListPoint,
) *Handler {
	return &Handler{
		ListPoints: ListPoints,
	}
}

// Routes registra las rutas HTTP disponibles para el mapa.
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
