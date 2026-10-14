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

	authapplication "metry.uk/provext4/MetroTrackerBackend/internal/auth/application"
	authpostgres "metry.uk/provext4/MetroTrackerBackend/internal/auth/infrastructure/postgres"
	authredis "metry.uk/provext4/MetroTrackerBackend/internal/auth/infrastructure/redis"
	authhttp "metry.uk/provext4/MetroTrackerBackend/internal/auth/transport/http"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/config"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/database"
	"metry.uk/provext4/MetroTrackerBackend/internal/platform/httpx"
	redisplatform "metry.uk/provext4/MetroTrackerBackend/internal/platform/redis"

	userapplication "metry.uk/provext4/MetroTrackerBackend/internal/users/application"
	userdomain "metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
	usergoogle "metry.uk/provext4/MetroTrackerBackend/internal/users/infrastructure/google"
	userjwt "metry.uk/provext4/MetroTrackerBackend/internal/users/infrastructure/jwt"
	userpostgres "metry.uk/provext4/MetroTrackerBackend/internal/users/infrastructure/postgres"
	userhttp "metry.uk/provext4/MetroTrackerBackend/internal/users/transport/http"

	mapaaplication "metry.uk/provext4/MetroTrackerBackend/internal/metroMap/application"
	mapapostgres "metry.uk/provext4/MetroTrackerBackend/internal/metroMap/infrastructure/postgres"
	maparedis "metry.uk/provext4/MetroTrackerBackend/internal/metroMap/infrastructure/redis"
	mapashttp "metry.uk/provext4/MetroTrackerBackend/internal/metroMap/trasport/http"
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

	tokenStore, err := authredis.NewTokenStore(redisClient, cfg.AuthTokenPrefix, cfg.AuthTokenTTL)
	if err != nil {
		log.Fatalf("auth token store configuration error: %v", err)
	}
	jwtIssuer, err := userjwt.NewIssuer(cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		log.Fatalf("JWT configuration error: %v", err)
	}
	refreshRepository := authpostgres.NewRefreshTokenRepository(dbPool)
	sessions := authapplication.NewSessionService(jwtIssuer, tokenStore, refreshRepository, cfg.JWTExpiration)

	// Estas capas conectan la persistencia con los casos de uso y el handler HTTP de usuarios.
	userRepository := userpostgres.NewUserRepository(
		dbPool,
	)

	createUser := userapplication.NewCreateUser(
		userRepository,
		sessions,
	)

	getUser := userapplication.NewGetUser(
		userRepository,
	)

	updateUser := userapplication.NewUpdateUser(
		userRepository,
	)

	deactivateUser := userapplication.NewDeactivateUser(
		userRepository,
	)

	listGenders := userapplication.NewListGenders(
		userRepository,
	)
	googleVerifier, err := usergoogle.NewIDTokenVerifier(cfg.GoogleClientID)
	if err != nil {
		log.Fatalf("Google verifier configuration error: %v", err)
	}
	googleLogin := userapplication.NewGoogleLogin(
		googleVerifier,
		userRepository,
		sessions,
		cfg.DefaultGenderID,
	)
	userHandler := userhttp.NewHandler(
		createUser,
		getUser,
		updateUser,
		deactivateUser,
		listGenders,
	)
	login := authapplication.NewLogin(userRepository, sessions, refreshRepository, cfg.JWTExpiration)
	refresh := authapplication.NewRefresh(userRepository, sessions, refreshRepository, cfg.JWTExpiration)
	authenticateToken := authapplication.NewAuthenticateToken(tokenStore, jwtIssuer, authUserReader{repository: userRepository})
	authMiddleware := authhttp.NewAuthMiddleware(authenticateToken)
	userHandler.SetAuthMiddleware(authMiddleware)
	authHandler := authhttp.NewHandler(googleLogin, authenticateToken, login, refresh)

	mapaRepository := mapapostgres.NewMapaRepository(
		dbPool,
	)

	mapaRepositoryWithCache := maparedis.NewMapaCacheRepository(
		mapaRepository,
		redisClient,
	)

	listPointsMap := mapaaplication.NewListPoint(
		mapaRepositoryWithCache,
	)

	mapasHandler := mapashttp.NewHandler(
		listPointsMap,
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
			// Auth se monta antes del router general para conservar explícitamente sus endpoints públicos.
			r.Mount("/auth", authHandler.Routes())
			// Las rutas versionadas de usuarios se montan bajo el handler de transporte.
			r.Mount("/users", userHandler.Routes())
			r.With(authMiddleware.Require).Mount("/maps", mapasHandler.Routes())
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

type authUserReader struct{ repository userdomain.UserRepository }

func (r authUserReader) AuthUser(ctx context.Context, id int32) (bool, error) {
	user, err := r.repository.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.EsActivo, nil
}
