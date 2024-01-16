package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"vivian.app/pkg/auth"
)

func EchoResponseHandler(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		echoResponse := vars["echo"]
		server.Logger.LogSuccess(echoResponse)
	})
}

func Authentication2FAHandler(ctx context.Context, server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resultChan := make(chan string)
		errChan := make(chan error)

		go func() {
			key2FA, err := auth.GenerateAuthKey2FA(ctx, server.Logger)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- key2FA
		}()

		select {
		case key2FA := <-resultChan:
			bytes, err := json.Marshal(key2FA)
			if err != nil {
				server.Logger.LogError("failure marshalling results", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
				server.Logger.LogError("failure writing results", err)
				return
			}
		case err := <-errChan:
			server.Logger.LogError("unable to generate authentication 2FA: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
