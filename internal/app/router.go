package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"vivian.app/internal/pkg/auth"
)

func EchoResponseHandler(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		echoResponse := vars["echo"]

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Logger.LogSuccess(echoResponse)
		}()
		wg.Wait()

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(echoResponse))
		if err != nil {
			server.Logger.LogError("Error writing response: ", err)
		}
	})
}

func Authentication2FA(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		action := strings.TrimSpace(q.Get("action"))
		switch action {
		case "generate":
			GenerateAuthentication2FA(w, ctx, server)
		case "verify":
			hash := strings.TrimSpace(q.Get("hash"))
			key := strings.TrimSpace(q.Get("key"))
			VerifyAuthentication2FA(w, ctx, server, hash, key)
		default:
			http.NotFound(w, r)
		}
	})
}

func GenerateAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server) {
	hashChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		hash2FA, err := auth.GenerateAuthKey2FA(ctx, server.Logger)
		if err != nil {
			errorChan <- err
			return
		}
		hashChan <- hash2FA
	}()

	select {
	case hash2FA := <-hashChan:
		bytes, err := json.Marshal(hash2FA)
		if err != nil {
			server.Logger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("failure writing results", err)
			return
		}
	case err := <-errorChan:
		server.Logger.LogError("unable to generate authentication 2FA: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func VerifyAuthentication2FA(w http.ResponseWriter, ctx context.Context, server *Server, hash2FA, key2FA string) {
	resultChan := make(chan bool)
	errorChan := make(chan error)

	go func() {
		result, err := auth.VerifyAuthKey2FA(ctx, hash2FA, key2FA, server.Logger)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		bytes, err := json.Marshal(result)
		if err != nil {
			server.Logger.LogError("failure marshalling results", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("failure writing results", err)
			return
		}
	case err := <-errorChan:
		//server.Logger.LogWarning("invalid key")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
