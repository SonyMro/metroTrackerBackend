package application

import (
	"context"
	"errors"
	"testing"

	"metry.uk/provext4/MetroTrackerBackend/internal/metroMap/domain"
)

// --- MOCK DEL REPOSITORIO ---
// Creamos un mock manual que implemente domain.MapaRepository
type MockMapaRepository struct {
	GetPointsFn func(ctx context.Context) ([]*domain.PuntoMap, error)
}

func (m *MockMapaRepository) GetPoints(ctx context.Context) ([]*domain.PuntoMap, error) {
	return m.GetPointsFn(ctx)
}

func TestListPointsExcute(t *testing.T) {
	ctx := context.Background()
	// Datos de prueba
	mockPoints := []*domain.PuntoMap{
		{
			Linea:      "Línea 1",
			Estacion:   "PANTITLÁN",
			Habilitada: 1,
			Longitud:   -99.0742379799485,
			Latitud:    19.4150858818997,
			Color:      "#f44a91",
		},
		{
			Linea:      "Línea 1",
			Estacion:   "OBSERVATORIO",
			Habilitada: 1,
			Longitud:   -99.2002533028881,
			Latitud:    19.3982046356116,
			Color:      "#f44a91",
		},
		{
			Linea:      "Línea 1",
			Estacion:   "JUANACATLÁN",
			Habilitada: 1,
			Longitud:   -99.1820020973682,
			Latitud:    19.4129381645968,
			Color:      "#f44a91",
		},
	}

	tests := []struct {
		name          string
		mockBehavior  func(ctx context.Context) ([]*domain.PuntoMap, error)
		expectedItems []*domain.PuntoMap
		expectedErr   string
	}{
		{
			name: "Exito: retorna lista de puntos",
			mockBehavior: func(ctx context.Context) ([]*domain.PuntoMap, error) {
				return mockPoints, nil
			},
			expectedItems: mockPoints,
			expectedErr:   "",
		},
		{
			name: "Error: repositorio falla",
			mockBehavior: func(ctx context.Context) ([]*domain.PuntoMap, error) {
				return nil, errors.New("db connection fail")
			},
			expectedItems: nil,
			expectedErr:   "Lista de puntos db connection fail",
		},
		{
			name: "Exito: lista vacia",
			mockBehavior: func(ctx context.Context) ([]*domain.PuntoMap, error) {
				return []*domain.PuntoMap{}, nil
			},
			expectedItems: []*domain.PuntoMap{},
			expectedErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup
			mockRepo := &MockMapaRepository{GetPointsFn: tt.mockBehavior}
			uc := NewListPoint(mockRepo)

			// 2. Ejecución
			result, err := uc.Execute(ctx)

			// 3. Verificación
			if tt.expectedErr != "" {
				if err == nil || err.Error() != tt.expectedErr {
					t.Errorf("se esperaba error %s, pero se obtuvo: %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("no se esperaba error, pero se obtuvo: %v", err)
				}
				if len(result) != len(tt.expectedItems) {
					t.Errorf("se esperaba %d items, pero se obtuvieron %d", len(tt.expectedItems), len(result))
				}
			}
		})
	}
}
