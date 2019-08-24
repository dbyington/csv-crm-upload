package sender

import (
	"bytes"
	"io"
	"net/rpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

var _ = Describe("Sender", func() {
	var (
		s *Sender

		rpcBuffer io.ReadWriteCloser
		rpcClient *rpc.Client
	)

	BeforeEach(func() {
		b := new(bytes.Buffer)
		rpcBuffer = &buffer{*b}
		rpcClient = rpc.NewClient(rpcBuffer)
		s = &Sender{client: rpcClient}
	})

	Context("NewSender", func() {
		var testS *Sender
		BeforeEach(func() {
			testS = NewSender(rpcClient)
		})

		It("should return a *Sender", func() {
			Expect(testS).To(BeEquivalentTo(s))
		})
	})

	// This test is flagged Pending because faking the *rpc.Client with a buffer just causes
	// any rpc.Client methods called to data race on the buffer and panic.
	PContext(".Close", func() {
		It("should close the client", func() {
			Expect(s.Close()).To(BeNil())
			Expect(s.Close()).To(MatchError(rpc.ErrShutdown))
		})
	})

	// This is the same as above, without a good way to really mock a *rpc.Client
	// this is also flagged Pending.
	PContext(".Signal", func() {
		It("should make the rpc call", func() {
			Expect(s.Signal()).ToNot(HaveOccurred())
		})
	})
})
