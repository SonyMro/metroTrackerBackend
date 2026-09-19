package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"metry.uk/provext4/MetroTrackerBackend/internal/platform/config"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/database"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
	redisplatform "metry.uk/provext4/MetroTrackerBackend/internal/platform/redis"

	userapplication "metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	userpostgres "metry.uk/provext4/MetroTrackerBackend/internal/users/infrastructure/postgres"
	userhttp "metry.uk/provext4/MetroTrackerBackend/internal/users/transport/http"
)

// main inicia la aplicación, conecta sus dependencias y levanta el servidor HTTP.
func main() {
	// La configuración se carga antes de componer las dependencias del proceso.
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// PostgreSQL se inicializa con los límites del pool y se libera al cerrar la aplicación.
	dbPool, err := database.NewPostgres(
		ctx,
		database.Config{
			URL:      cfg.DatabaseURL,
			MaxConns: cfg.DBMaxConns,
			MinConns: cfg.DBMinConns,
		},
	)

	if err != nil {
		log.Fatalf(
			"database connection error: %v",
			err,
		)
	}

	defer dbPool.Close()

	// Redis se crea y se comprueba antes de exponer las rutas que dependen de él.
	redisClient, err := redisplatform.NewClient(
		cfg.RedisURL,
	)

	if err != nil {
		log.Fatalf(
			"redis configuration error: %v",
			err,
		)
	}

	defer redisClient.Close()

	if err := redisplatform.Ping(
		ctx,
		redisClient,
	); err != nil {
		log.Fatalf(
			"redis connection error: %v",
			err,
		)
	}

	// Estas capas conectan la persistencia con los casos de uso y el handler HTTP de usuarios.
	userRepository := userpostgres.NewUserRepository(
		dbPool,
	)

	createUser := userapplication.NewCreateUser(
		userRepository,
	)

	getUser := userapplication.NewGetUser(
		userRepository,
	)

	listUsers := userapplication.NewListUsers(
		userRepository,
	)

	updateUser := userapplication.NewUpdateUser(
		userRepository,
	)

	deactivateUser := userapplication.NewDeactivateUser(
		userRepository,
	)

	userHandler := userhttp.NewHandler(
		createUser,
		getUser,
		listUsers,
		updateUser,
		deactivateUser,
	)

	router := chi.NewRouter()

	// Los middlewares añaden identificación, dirección real, registro, recuperación y límite de tiempo.
	router.Use(
		middleware.RequestID,
	)

	router.Use(
		middleware.RealIP,
	)

	router.Use(
		middleware.Logger,
	)

	router.Use(
		middleware.Recoverer,
	)

	router.Use(
		middleware.Timeout(30 * time.Second),
	)

	router.Get(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			// Liveness: confirma que el proceso HTTP sigue atendiendo solicitudes.
			httpx.WriteJSON(
				w,
				http.StatusOK,
				map[string]string{
					"status": "ok",
				},
			)
		},
	)

	router.Get(
		"/health/ready",
		func(w http.ResponseWriter, r *http.Request) {
			// Readiness exige que las dependencias requeridas estén disponibles.
			if err := dbPool.Ping(
				r.Context(),
			); err != nil {
				httpx.WriteError(
					w,
					http.StatusServiceUnavailable,
					"database unavailable",
				)

				return
			}

			if err := redisplatform.Ping(
				r.Context(),
				redisClient,
			); err != nil {
				httpx.WriteError(
					w,
					http.StatusServiceUnavailable,
					"redis unavailable",
				)

				return
			}

			httpx.WriteJSON(
				w,
				http.StatusOK,
				map[string]string{
					"status": "ready",
				},
			)
		},
	)

	router.Route(
		"/api/v1",
		func(r chi.Router) {
			// Las rutas versionadas de usuarios se montan bajo el handler de transporte.
			r.Mount(
				"/users",
				userHandler.Routes(),
			)
		},
	)

	// Estos límites protegen la lectura, escritura y permanencia de conexiones HTTP.
	server := &http.Server{
		Addr: cfg.HTTPAddr,

		Handler: router,

		ReadHeaderTimeout: 5 * time.Second,

		ReadTimeout: 10 * time.Second,

		WriteTimeout: 30 * time.Second,

		IdleTimeout: 60 * time.Second,
	}

	serverErrors := make(
		chan error,
		1,
	)

	// El servidor arranca en segundo plano para poder atender errores y señales de apagado.
	go func() {
		log.Printf(
			"HTTP server listening on %s",
			cfg.HTTPAddr,
		)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	// Se espera el cierre del servidor o una señal del sistema para iniciar un apagado ordenado.
	select {
	case err := <-serverErrors:
		if !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			log.Fatalf(
				"HTTP server error: %v",
				err,
			)
		}

	case <-shutdownContext.Done():
		log.Println(
			"shutdown signal received",
		)
	}

	// El límite permite terminar solicitudes en curso sin bloquear el cierre indefinidamente.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {
		log.Printf(
			"HTTP server shutdown error: %v",
			err,
		)
	}
}
