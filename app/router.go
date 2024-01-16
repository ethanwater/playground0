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
		key2FA, err := auth.GenerateAuthKey2FA(ctx, server.Logger)
		if err != nil {
			server.Logger.LogError("unable to generate authentication 2FA")
			return
		}

		bytes, err := json.Marshal(key2FA)
		if err != nil {
			server.Logger.LogError("failure marshalling results")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
			server.Logger.LogError("failure writing results")
			return
		}
	})
}
