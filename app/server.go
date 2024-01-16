package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	_ "embed"

	"github.com/TwiN/go-color"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"vivian.app/utils"
)

const (
	VivianAppName      string        = "vivian.app"
	VivianHostAddr     string        = ":8080"
	VivianReadTimeout  time.Duration = time.Second * 10
	VivianWriteTimeout time.Duration = time.Second * 10
)

type ServerInitialization interface {
	Deploy(context.Context) error
}

type Server struct {
	DeploymentID       string
	Listener           net.Listener
	Handler            http.Handler
	Logger             *utils.VivianLogger
	Addr               string
	VivianReadTimeout  time.Duration
	VivianWriteTimeout time.Duration
	mux                sync.Mutex
}

func Deploy(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.Lmsgprefix)
	s := buildServer(ctx, logger)
	s.mux.Lock()
	defer s.mux.Unlock()

	server := &http.Server{
		Addr:         s.Addr,
		Handler:      s.Handler,
		ReadTimeout:  s.VivianReadTimeout,
		WriteTimeout: s.VivianWriteTimeout,
	}

	logDeployment := func(s *Server) {
		fmt.Printf("╭───────────────────────────────────────────────────╮\n")
		fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Purple, VivianAppName))
		fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Blue, s.DeploymentID))
		fmt.Printf("╰───────────────────────────────────────────────────╯\n")
	}

	logDeployment(s)

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return http.ListenAndServe(s.Addr, s.Handler)
}

func buildServer(ctx context.Context, logger *log.Logger) *Server {
	generateDeploymentID := func() string {
		randomUUID := uuid.New()
		shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
			randomUUID[8:10], randomUUID[10:])

		return shortUUID
	}

	deploymentID := generateDeploymentID()
	router := mux.NewRouter()

	server := &Server{
		DeploymentID:       deploymentID,
		Logger:             &utils.VivianLogger{Logger: logger, DeploymentID: deploymentID},
		Handler:            router,
		Addr:               VivianHostAddr,
		VivianReadTimeout:  VivianReadTimeout,
		VivianWriteTimeout: VivianWriteTimeout,
	}

	router.Handle("/echo={echo}", EchoResponseHandler(ctx, server))
	router.Handle("/user/{username}/2fa", Authentication2FAHandler(ctx, server))

	return server
}
