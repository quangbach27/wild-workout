package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	firebase "firebase.google.com/go"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/quangbach27/wild-workout/internal/common/auth"
	"github.com/quangbach27/wild-workout/internal/common/logs"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	httpEndpoint := fmt.Sprintf(":%s", port)

	apiRouter := chi.NewRouter()
	setMiddlewares(apiRouter)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", createHandler(apiRouter))

	logrus.Info("Starting HTTP server at:", httpEndpoint)
	http.ListenAndServe(httpEndpoint, rootRouter)
}

func setMiddlewares(router *chi.Mux) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logs.NewStructuredLogger(logrus.StandardLogger()))
	router.Use(middleware.Recoverer)

	addCorsMiddleware(router)
	addAuthMiddleware(router)

	router.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)
	router.Use(middleware.NoCache)
}

func addAuthMiddleware(router *chi.Mux) {
	if mockAuth, _ := strconv.ParseBool(os.Getenv("MOCK_AUTH")); mockAuth {
		router.Use(auth.HttpMockMiddleware)
		return
	}

	var opts []option.ClientOption
	if file := os.Getenv("SERVICE_ACCOUNT_FILE"); file != "" {
		opts = append(opts, option.WithCredentialsFile(file))
	}

	config := &firebase.Config{ProjectID: os.Getenv("GCP_PROJECT")}
	firebaseApp, err := firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		logrus.Fatalf("error initializing app: %v\n", err)
	}

	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create firebase Auth client")
	}

	router.Use(auth.FirebaseHTTPMiddleware{AuthClient: authClient}.Middleware)
}

func addCorsMiddleware(router *chi.Mux) {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
	if len(allowedOrigins) == 0 {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(corsMiddleware.Handler)
}
