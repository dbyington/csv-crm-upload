package main

import (
    "net/http"
)

type crm struct {}

func (c *crm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(""))
}

func main() {
    s := &http.Server{Addr: ":80"}
    http.Handle("/", &crm{})
    s.ListenAndServe()
}
