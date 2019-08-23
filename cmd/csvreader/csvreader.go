package csvreader

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/dbyington/csv-crm-upload/database"
)

type reader struct {
	*csv.Reader
	db         database.CustomerDB
	headerRow  bool
	lineBuffer int
}

func NewReader(db database.CustomerDB, f io.Reader, headerRow bool, lineBuffer int) *reader {
	return &reader{
		Reader:     csv.NewReader(f),
		db:         db,
		headerRow:  headerRow,
		lineBuffer: lineBuffer,
	}
}

func (c *reader) Run() error {
    if c.headerRow {
        if err := c.dropHeaderRow(); err != nil {
            return err
        }
    }
    if err := c.readCustomers(); err != io.EOF {
        return err
    }
    return nil
}

func (c *reader) dropHeaderRow() error {
	_, err := c.Read()
	return err
}

func (c *reader) readCustomers() error {
    if c.headerRow {
        if err := c.dropHeaderRow(); err != nil {
            return err
        }
    }
	customers := database.NewCustomers()
	for {
		if customers.Count() == c.lineBuffer {
			// insertCustomers will log any issues with inserting customers into the database. Once
			// row(s) have been inserted it will handle signalling to the CRM worker there are
			// customers ready to upload.
			insertCustomers(customers)

			// Clear the customers to start fresh.
			customers = database.NewCustomers()
		}

		// While we have rows to read parse them and append them to the customer set.
		if id, first, last, email, phone, err := c.parseRow(); err == nil {
			customers.Append(c.db.NewCustomer(id, first, last, email, phone))
		} else {
			if err == io.EOF {
				insertCustomers(customers)
				return err
			} else {
			    // If parseRow returns an error other than EOF just log it and continue.
			    log.Printf("error reading row: %s", err)
            }
		}
	}
}

func insertCustomers(customers database.Customers) {
	// Insert our customer set. If we get an error it will have failed on the entire set so range through the Customer
	// and try to insert the individual customers, logging which one(s) still fail.
	if err := customers.Insert(); err != nil {
		// TODO: Flesh out this messaging a bit more it needs to be descriptive about breaking down and trying
		//  individual customers.
		log.Print(err)

		for _, c := range customers.List() {
			if err := c.Insert(); err != nil {
				// TODO: This message needs to include the full customer row, if the db error is not explicit enough.
				log.Printf("ERROR inserting customer: %s", err)
			} else {
				// TODO: Signal row has been inserted.
			}
		}
	} else {
		// TODO: Signal rows have been inserted.
	}
}

func (c *reader) parseRow() (int64, string, string, string, string, error) {
	row, err := c.Read()
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
	return int64(id), row[1], row[2], row[3], row[4], nil
}
