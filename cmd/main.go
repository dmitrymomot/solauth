package main

import (
	"context"

	"github.com/dmitrymomot/solauth"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init logger
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"app":       appName,
		"build_tag": buildTagRuntime,
	})

	// set up jwt interactor
	jwtInteractor := solauth.NewJWT(authSigningKey)

	// Init HTTP router
	r := initRouter()

	// Endpoints
	r.Post("/auth/request", solauth.RequestAuth)
	r.Post("/auth/verify", solauth.VerifySignedMessage(jwtInteractor))
	r.Post("/auth/refresh", solauth.RefreshToken(jwtInteractor))

	// Run HTTP server
	runServer(httpPort, r, logger)
}
