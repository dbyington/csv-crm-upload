package csvreader

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"strconv"

	"github.com/dbyington/csv-crm-upload/database"
	"github.com/dbyington/csv-crm-upload/signal/sender"
)

type reader struct {
	*csv.Reader
	db         database.CustomerDB
	headerRow  bool
	bufferSize int
	sender     *sender.Sender
}

func NewReader(db database.CustomerDB, f io.Reader, c *rpc.Client, noHeaderRow bool, lineBuffer int) *reader {
	rpcSender := sender.NewSender(c)

	return &reader{
		Reader:     csv.NewReader(f),
		db:         db,
		headerRow:  !noHeaderRow,
		bufferSize: lineBuffer,
		sender:     rpcSender,
	}
}

func (r *reader) Run() error {
	defer r.sender.Close()

	if r.headerRow {
		if err := r.dropHeaderRow(); err != nil {
			return err
		}
	}
	if err := r.readCustomers(); err != io.EOF {
		return err
	}
	return nil
}

func (r *reader) dropHeaderRow() error {
	_, err := r.Read()
	return err
}

func (r *reader) readCustomers() error {
	if r.headerRow {
		if err := r.dropHeaderRow(); err != nil {
			return err
		}
	}
	customers := database.NewCustomers()
	for {
		if customers.Count() == r.bufferSize {
			// insertCustomers will log any issues with inserting customers into the database. Once
			// row(s) have been inserted it will handle signalling to the CRM worker there are
			// customers ready to upload.
			r.insertCustomers(customers)

			// Clear the customers to start fresh.
			customers = database.NewCustomers()
		}

		// While we have rows to read parse them and append them to the customer set.
		if id, first, last, email, phone, err := r.parseRow(); err == nil {
			customers.Append(r.db.NewCustomer(id, first, last, email, phone))
		} else {
			if err == io.EOF {
				r.insertCustomers(customers)
				return err
			} else {
				// If parseRow returns an error other than EOF just log it and continue.
				log.Printf("error reading row: %s", err)
			}
		}
	}
}

func (r *reader) insertCustomers(customers database.Customers) {
	// Insert our customer set. If we get an error it will have failed on the entire set so range through the Customer
	// and try to insert the individual customers, logging which one(s) still fail.
	if err := customers.Insert(); err != nil {
		log.Print("error while inserting customer set, trying individual customer inserts.")

		for _, c := range customers.List() {
			if err := c.Insert(); err != nil {
				log.Printf("ERROR inserting customer (%s %s, %s): %s", c.FirstName, c.LastName, c.Email, err)
			} else {
				if err := r.sender.Signal(); err != nil {
					log.Printf("ERROR signaling CRM after inserting new customers: %s", err)
				}
			}
		}
	} else {
		if err := r.sender.Signal(); err != nil {
			log.Printf("ERROR signaling CRM after inserting new customers: %s", err)
		}
	}
}

func (r *reader) parseRow() (int64, string, string, string, string, error) {
	row, err := r.Read()
	if err != nil {
		if parseErr, ok := err.(*csv.ParseError); ok {
			return 0, "", "", "", "", fmt.Errorf("parse error while reading row: %s", parseErr)
		} else if err == io.EOF {
			return 0, "", "", "", "", err
		}
		return 0, "", "", "", "", fmt.Errorf("error while reading row: %s", err)

	}
	id, err := strconv.Atoi(row[0])
	if err != nil {
		return 0, "", "", "", "", fmt.Errorf("failed to parse row id: %s", err)
	}

	if row[3] == "" {
		return 0, "", "", "", "", fmt.Errorf("failed to parse row: email cannot be empty")
	}
	return int64(id), row[1], row[2], row[3], row[4], nil
}
