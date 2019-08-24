package sender

import (
	"net/rpc"
)

type Sender struct {
	client *rpc.Client
}

func NewSender(c *rpc.Client) *Sender {
	return &Sender{client: c}
}

func (s *Sender) Signal() error {
	rec := &struct{}{}
	if err := s.client.Call("Signaler.Send", struct{}{}, rec); err != nil {
		return err
	}

	return nil
}

func (s *Sender) Close() error {
	return s.client.Close()
}
