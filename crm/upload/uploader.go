package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dbyington/csv-crm-upload/database"
	"github.com/dbyington/csv-crm-upload/signal/listener"
)

// The maximum number of upload routines to spawn at one time.
const maxConcurrentUploads = 25

// The maximum time to wait for the CRM Server.
const clientTimeout = 30

type upload struct {
	listenAddress    string
	crmServerAddress string
	crmAPI           string
	httpClient       *http.Client
	db               database.CustomerDB
	sigChan          chan struct{}
	successChan      chan struct{}
	stopRun          context.CancelFunc
	closeQueue       context.CancelFunc
	wg               sync.WaitGroup
	uploadChan       chan database.Customer
}

func NewUploader(lis, crm, crmAPI string, db database.CustomerDB) *upload {
	return &upload{
		listenAddress:    lis,
		crmServerAddress: crm,
		crmAPI:           crmAPI,
		db:               db,
		httpClient: &http.Client{
			Timeout: clientTimeout * time.Second,
		},
		sigChan:     make(chan struct{}, 1),
		uploadChan:  make(chan database.Customer, maxConcurrentUploads),
		successChan: make(chan struct{}, 1),
	}
}

// Start starts the uploader service.
func (u *upload) Start() {
	l := listener.NewListener(u.listenAddress, u.sigChan)
	ctxRun, cancelRun := context.WithCancel(context.Background())
	ctxQueue, cancelQueue := context.WithCancel(context.Background())
	u.stopRun = cancelRun
	u.closeQueue = cancelQueue
	go u.run(ctxRun)
	go u.uploadQueue(ctxQueue)
	log.Fatal(l.Start())
}

// Stop will signal the running uploader go routines to finish and return then wait for any other processes to finish.
func (u *upload) Stop() {
	u.stopRun()    // Signals the run() loops that we're Stop has been called.
	u.closeQueue() // Signals the upload queue to finish and exit.
	u.wg.Wait()    // wait for any inflight work to finish
}

func (u *upload) run(ctx context.Context) {
	fib := fibFunc()
	timer := time.NewTimer(time.Second)

	for {
		select {
		// If the last upload was a success then reset the fib sequence
		case <-u.successChan:
			fib = fibFunc()
			// The empty default will cause this select to fall through to the next one and not wait for success.
		default:
		}

		select {
		case <-u.sigChan:
			u.processNewCustomers()
			timer.Reset(time.Duration(fib()) * time.Second)
		case <-ctx.Done():
			log.Print("we're done here")
			return
		case <-timer.C:
			log.Print("checking for work")
			u.processNewCustomers()
			timer.Reset(time.Duration(fib()) * time.Second)
		}
	}
}

func (u *upload) processNewCustomers() {
	customers, err := u.db.SelectCustomersForUpload()
	if err != nil {
		log.Print(fmt.Errorf("error getting new customers for upload: %s", err))
	}

	if customers.Count() == 0 {
		return
	}

	log.Printf("Processing %d customers", customers.Count())
	for _, customer := range customers.List() {
		u.uploadChan <- customer
	}
	log.Print("done.")
}

func (u *upload) post(c database.Customer) error {
	customerJSON, err := json.Marshal(c)
	req := bytes.NewBuffer(customerJSON)
	if err != nil {
		return fmt.Errorf("error marshaling customerr: %s", err)
	}

	resp, err := u.httpClient.Post(u.crmServerAddress+u.crmAPI, "application/json", req)
	if err != nil {
		return fmt.Errorf("error while posting to CRM: %s", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("post to CRM failed with (%d) %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func (u *upload) uploadQueue(ctx context.Context) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case customer := <-u.uploadChan:
			if err = u.post(customer); err == nil {
				if err = customer.Uploaded(); err == nil {
					u.success()
				}
			}
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func (u *upload) success() {
	select {
	case u.successChan <- struct{}{}:
	default:
	}
}

// fibFunc is a helper function to return a function that produces a fibonacci sequence, returning the next number with
// each successive call.
func fibFunc() func() int {
	var a, b int
	a = 1
	return func() int {
		c := a + b
		b = a
		a = c
		return c
	}
}
