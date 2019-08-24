package listener

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/rpc"
)

type Listener struct {
	addr     string
	signaler *Signaler
	server   *http.Server
}

var errBusy = errors.New("signal listener busy")

type Signaler struct {
	sig chan<- struct{}
}

func NewListener(a string, s chan<- struct{}) *Listener {
	return &Listener{
		addr:     a,
		server:   &http.Server{Addr: a},
		signaler: &Signaler{sig: s},
	}
}

func (s *Signaler) Send(args, reply *struct{}) error {
	// Send in a select so that if the channel is not ready to receive, i.e is buffered but full, this will not block.
	select {
	case s.sig <- struct{}{}:
		return nil
	default:
		return errBusy
	}
}

func (l *Listener) Start() error {
	// Any error returned here is not recoverable since it is programmatic.
	if err := rpc.Register(l.signaler); err != nil {
		log.Fatalf("while registering rpc: %s", err)
	}
	rpc.HandleHTTP()
	return l.server.ListenAndServe()
}

func (l *Listener) Stop() {
	// Not really concerned about errors during shutdown.
	_ = l.server.Shutdown(context.Background())
}
