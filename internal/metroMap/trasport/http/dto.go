package http

import "metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"

type PointMapResponse struct {
	Linea      string  `json:"linea"`
	Estacion   string  `json:"estacion"`
	Habilitada int     `json:"habilitada"`
	Longitud   float64 `json:"longitud"`
	Latitud    float64 `json:"latitud"`
	Color      string  `json:"color"`
}

func toMapsResponse(
	punto *domain.PuntoMap,
) *PointMapResponse {
	response := &PointMapResponse{
		Linea:      punto.Linea,
		Estacion:   punto.Estacion,
		Habilitada: punto.Habilitada,
		Longitud:   punto.Longitud,
		Latitud:    punto.Latitud,
		Color:      punto.Color,
	}

	return response
}
