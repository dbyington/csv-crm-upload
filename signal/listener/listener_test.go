package listener

import (
	"context"
	"net/http"
	"net/rpc"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const addr = ":9876"

var _ = Describe("Listener", func() {
	var (
		err     error
		l       *Listener
		s       *Signaler
		sigChan chan struct{}
		server  *http.Server
		testL   *Listener
	)

	BeforeEach(func() {
		server = &http.Server{Addr: addr}
		sigChan = make(chan struct{}, 1)
		s = &Signaler{sig: sigChan}

		l = &Listener{
			addr:     addr,
			signaler: s,
			server:   server,
		}
	})

	Context("NewListener", func() {
		BeforeEach(func() {
			testL = NewListener(addr, sigChan)
			testL.signaler.sig <- struct{}{}
		})

		It("should return a Listener", func() {
			Expect(testL).ToNot(BeNil())
			Expect(testL).To(BeEquivalentTo(l))
			Expect(sigChan).To(Receive())
		})
	})

	Context(".Start", func() {
		BeforeEach(func() {
			go l.Start()
		})

		AfterEach(func() {
			_ = l.server.Shutdown(context.Background())
		})

		Context("with a good signaler", func() {
			It("should register the signaler", func() {
				_, err := rpc.DialHTTP("tcp", "localhost"+addr)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context(".Stop", func() {
		var shutdown int32
		BeforeEach(func() {
			l.server.RegisterOnShutdown(func() { atomic.AddInt32(&shutdown, 1) })
			go l.server.ListenAndServe()

			// Give the server a moment to start
			time.Sleep(10 * time.Millisecond)

			l.Stop()
		})

		Context("when called", func() {
			It("should stop the server", func() {
				Expect(atomic.LoadInt32(&shutdown)).To(Equal(int32(1)))
			})
		})
	})

	Context("Signaler", func() {
		Context(".Send", func() {
			Context("when the channel is not blocked", func() {
				BeforeEach(func() {
					err = s.Send(&struct{}{}, &struct{}{})
				})

				It("should send on the channel", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(sigChan).To(Receive())
				})
			})

			Context("when the channel is full", func() {
				BeforeEach(func() {
					sigChan <- struct{}{}
					err = s.Send(&struct{}{}, &struct{}{})
				})

				It("should not be blocked", func() {
					Expect(err).To(MatchError(errBusy))
				})
			})
		})
	})
})
