package app

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

func EchoResponseHandler(ctx context.Context, server *T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		echoResponse := vars["echo"]
		server.LogSuccess(echoResponse)
	})
}
