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
)

const VivianAppName string = "vivian.app"

type Server interface {
	Deploy(context.Context) error
}

type T struct {
	DeploymentID string
	Listener     net.Listener
	Handler      http.Handler
	Logger       *log.Logger
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	mux          sync.Mutex
}

func (t *T) LogFatal(msg string) {
	t.Logger.SetPrefix(fmt.Sprintf("%s	%s", color.Ize(color.Purple, t.DeploymentID[:8]), color.Ize(color.RedBackground, "[vivian:fatal]")))
	t.Logger.Printf("%+8s", msg)
	os.Exit(1)
}

func (t *T) LogWarning(msg string) {
	t.Logger.SetPrefix(fmt.Sprintf("%s	%s", color.Ize(color.Purple, t.DeploymentID[:8]), color.Ize(color.Yellow, "[vivian:warn]")))
	t.Logger.Printf("%+9s", msg)
}

func (t *T) LogError(msg string) {
	t.Logger.SetPrefix(fmt.Sprintf("%s	%s", color.Ize(color.Purple, t.DeploymentID[:8]), color.Ize(color.Red, "[vivian:error]")))
	t.Logger.Printf("%+8s", msg)
}

func (t *T) LogSuccess(msg string) {
	t.Logger.SetPrefix(fmt.Sprintf("%s	%s", color.Ize(color.Purple, t.DeploymentID[:8]), color.Ize(color.Green, "[vivian:success]")))
	t.Logger.Printf("%+6s", msg)
}

func Deploy(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC|log.Lmsgprefix)
	t := buildServer(ctx, logger)
	t.mux.Lock()
	defer t.mux.Unlock()

	server := &http.Server{
		Addr:         t.Addr,
		Handler:      t.Handler,
		ReadTimeout:  t.ReadTimeout,
		WriteTimeout: t.WriteTimeout,
	}

	displayDeployment := func(t *T) {
		fmt.Printf("╭───────────────────────────────────────────────────╮\n")
		fmt.Printf("│ app        : %-45s │\n", color.Ize(color.Purple, VivianAppName))
		fmt.Printf("│ deployment : %-36s │\n", color.Ize(color.Blue, t.DeploymentID))
		fmt.Printf("╰───────────────────────────────────────────────────╯\n")
	}

	displayDeployment(t)

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return http.ListenAndServe(t.Addr, t.Handler)
}

func buildServer(ctx context.Context, logger *log.Logger) *T {
	generateDeploymentID := func() string {
		randomUUID := uuid.New()
		shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
			randomUUID[8:10], randomUUID[10:])

		return shortUUID
	}

	deploymentID := generateDeploymentID()
	router := mux.NewRouter()

	server := &T{
		DeploymentID: deploymentID,
		Logger:       logger,
		Handler:      router,
		Addr:         ":8080",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	router.HandleFunc("/{username}", func(w http.ResponseWriter, r *http.Request) {
		server.LogSuccess("neko")
	})

	return server
}
