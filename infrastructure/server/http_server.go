package server

import (
    "net/http"
    "time"
)

func NewHTTPServer(handler http.Handler) *http.Server {
    return &http.Server{
        Addr:              ":8080",
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
        IdleTimeout:       60 * time.Second,
    }
}