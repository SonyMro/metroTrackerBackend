package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/application"
	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
)

type mapaRepositoryStub struct {
	points []*domain.PuntoMap
	err    error
}

func (stub *mapaRepositoryStub) GetPoints(context.Context) ([]*domain.PuntoMap, error) {
	return stub.points, stub.err
}

func TestHandlerGetDatosMapa(t *testing.T) {
	points := []*domain.PuntoMap{{
		Linea:      "Línea 1",
		Estacion:   "PANTITLÁN",
		Habilitada: 1,
		Longitud:   -99.0742,
		Latitud:    19.415,
		Color:      "#f44a91",
	}}

	tests := []struct {
		name          string
		points        []*domain.PuntoMap
		repositoryErr error
		wantStatus    int
		wantBody      string
	}{
		{
			name:       "éxito con puntos",
			points:     points,
			wantStatus: http.StatusOK,
			wantBody:   `[{"Linea":"Línea 1","Estacion":"PANTITLÁN","Habilitada":1,"Longitud":-99.0742,"Latitud":19.415,"Color":"#f44a91"}]` + "\n",
		},
		{
			name:          "error de dominio",
			repositoryErr: domain.ErrorPunto,
			wantStatus:    http.StatusFound,
			wantBody:      `{"error":"punto no funciona"}` + "\n",
		},
		{
			name:          "error interno",
			repositoryErr: errors.New("fallo inesperado"),
			wantStatus:    http.StatusInternalServerError,
			wantBody:      `{"error":"internla server error"}` + "\n",
		},
		{
			name:       "lista vacía",
			points:     []*domain.PuntoMap{},
			wantStatus: http.StatusOK,
			wantBody:   "[]\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &mapaRepositoryStub{
				points: test.points,
				err:    test.repositoryErr,
			}
			useCase := application.NewListPoint(repository)
			handler := NewHandler(useCase)
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			response := httptest.NewRecorder()

			handler.Routes().ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, se esperaba %d", response.Code, test.wantStatus)
			}
			if got := response.Body.String(); got != test.wantBody {
				t.Errorf("body = %q, se esperaba %q", got, test.wantBody)
			}
			if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Errorf("Content-Type = %q, se esperaba application/json", contentType)
			}
		})
	}
}
