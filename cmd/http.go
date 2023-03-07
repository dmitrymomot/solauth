package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type (
	// logger interface
	logger interface {
		Infof(format string, args ...interface{})
		Fatalf(format string, args ...interface{})
	}
)

// Init HTTP router
func initRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Recoverer,
		middleware.Logger,
		middleware.Timeout(httpRequestTimeout),
		middleware.CleanPath,
		middleware.StripSlashes,
		middleware.GetHead,
		middleware.NoCache,
		middleware.RealIP,
		middleware.RequestID,

		middleware.AllowContentType(
			"application/json",
			"multipart/form-data",
			"application/x-www-form-urlencoded",
		),

		// Rate limit by IP address.
		httprate.LimitByIP(10, 1*time.Minute),

		// Basic CORS
		// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
		cors.Handler(cors.Options{
			AllowedOrigins:   corsAllowedOrigins,
			AllowedMethods:   corsAllowedMethods,
			AllowedHeaders:   corsAllowedHeaders,
			AllowCredentials: corsAllowedCredentials,
			MaxAge:           corsMaxAge, // Maximum value not ignored by any of major browsers
		}),

		// Uses for testing error response with needed status code
		testingMdw,
	)

	r.NotFound(notFoundHandler)
	r.MethodNotAllowed(methodNotAllowedHandler)

	r.Get("/", mkRootHandler(buildTagRuntime))
	r.Get("/health", healthCheckHandler)

	return r
}

// Run HTTP server
func runServer(httpPort int, router http.Handler, log logger) {
	log.Infof("Starting HTTP server on port %d", httpPort)

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	defer serverStopCtx()

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGPIPE)

	httpServer := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", httpPort),
	}

	go func() {
		<-sig
		log.Infof("Received signal to stop HTTP server")

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(serverCtx, httpServerShutdownTimeout)
		defer shutdownCtxCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Infof("HTTP server shutdown timeout exceeded, force shutdown")
				os.Exit(1)
			}
		}()

		// Trigger graceful shutdown
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("HTTP server shutdown error: %s", err)
		}
		serverStopCtx()
	}()

	// Run the server
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %s", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	log.Infof("HTTP server stopped")
}

// returns 204 HTTP status without content
func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// returns 404 HTTP status with payload
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	defaultResponse(w, http.StatusNotFound, map[string]interface{}{
		"code":       http.StatusNotFound,
		"error":      fmt.Sprintf("Endpoint %s", http.StatusText(http.StatusNotFound)),
		"request_id": middleware.GetReqID(r.Context()),
	})
}

// returns 405 HTTP status with payload
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	defaultResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{
		"code":       http.StatusMethodNotAllowed,
		"error":      http.StatusText(http.StatusMethodNotAllowed),
		"request_id": middleware.GetReqID(r.Context()),
	})
}

// returns current build tag
func mkRootHandler(buildTag string) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		defaultResponse(w, http.StatusOK, map[string]interface{}{
			"build_tag": buildTag,
		})
	}
}

// helper to send response as a json data
func defaultResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Testing middleware
// Helps to test any HTTP error
// Pass must_err query parameter with code you want get
// E.g.: /login?must_err=403
func testingMdw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errCodeStr := r.URL.Query().Get("must_err"); len(errCodeStr) == 3 {
			if errCode, err := strconv.Atoi(errCodeStr); err == nil && errCode >= 400 && errCode < 600 {
				defaultResponse(w, errCode, map[string]interface{}{
					"code":       errCode,
					"error":      http.StatusText(errCode),
					"request_id": middleware.GetReqID(r.Context()),
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
