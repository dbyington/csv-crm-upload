package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	// This external lib is required for postgres.
	_ "github.com/lib/pq"
)

const (
	insertTemplate      = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORD(null::customers, '{{.JSON}}');`
	insertSetTemplate   = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORDSET(null::customers, {{.JSON}});`
	selectUploadedFalse = `SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false;`
	updateUploaded      = `UPDATE customers SET uploaded = true WHERE email = $1;`
)

var (
	db                        *sql.DB
	insertCustomerTemplate    *template.Template
	insertCustomerSetTemplate *template.Template
)

// Customer describes a CRM customer
type customer struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type Customer interface {
    Insert() error
    Uploaded() error
}

// customers is a slice of *Customer
type customers []*customer

type Customers interface {
    Append(*customer)
    Count() int
    Insert() error
    List() []*customer
}

type templateFields struct {
	JSON string
}

func init() {
	// These only need to be run once and they don't produce errors so this is a good place for them.
	insertCustomerTemplate = template.Must(template.New("insertCustomer").Parse(insertTemplate))
	insertCustomerSetTemplate = template.Must(template.New("insertCustomerSet").Parse(insertSetTemplate))
}

// Open creates the database connection to the postgres database.
func Open(user, password, host, database string) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, database)
	d, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("while opening database: %s", err)
	}
	db = d
}

// NewCustomer returns a *Customer based on the supplied Customer type values.
func NewCustomer(id int64, firstName, lastName, email, phone string) *customer {
	return &customer{
		id,
		firstName,
		lastName,
		email,
		phone,
	}
}

// NewCustomers creates a *customers object consisting of the optionally supplied *Customer objects.
func NewCustomers(customerList ...*customer) *customers {
	customers := &customers{}
	for _, customer := range customerList {
		*customers = append(*customers, customer)
	}

	return customers
}

// Insert performs the database insert of a single customer object.
func (c *customer) Insert() error {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("while marshaling customer: %s", err)
	}

	s := &templateFields{JSON: string(jsonBytes)}
	b := new(bytes.Buffer)

	err = insertCustomerTemplate.Execute(b, s)
	if err != nil {
		return fmt.Errorf("customer data: %s", err)
	}

	return insert(b.Bytes())
}

// Append adds the supplied *Customer to the *customers set.
func (c *customers) Append(customer *customer) {
    *c = append(*c, customer)
}

// Count returns the number of *Customer in the *customers set.
func (c *customers) Count() int {
    return len(*c)
}

func (c *customers) List() []*customer {
    return *c
}

// Insert performs the database insert with a set of customer objects
func (c *customers) Insert() error {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("while marshaling customer: %s", err)
	}

	s := &templateFields{JSON: string(jsonBytes)}
	b := new(bytes.Buffer)

	err = insertCustomerSetTemplate.Execute(b, s)
	if err != nil {
		return fmt.Errorf("customer data: %s", err)
	}

	return insert(b.Bytes())
}

func insert(b []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("while creating transaction: %s", err)
	}

	// If returning an error Rollback the transaction, otherwise Commit it.
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	_, err = tx.Exec(string(b))
	if err != nil {
		return fmt.Errorf("executing query: %s", err)
	}

	return nil
}

// SelectCustomersForUpload returns a *customers struct suitable to be unmarshaled and uploaded to CRM.
func SelectCustomersForUpload() (*customers, error) {
	customers := new(customers)

	rows, err := db.Query(selectUploadedFalse)
	if err != nil {
		return nil, fmt.Errorf("while selecting rows: %s", err)
	}

	for rows.Next() {
		c := new(customer)
		err := rows.Scan(&c.Id, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
		if err != nil {
			return nil, fmt.Errorf("while scanning rows: %s", err)
		}

		*customers = append(*customers, c)
	}

	return customers, nil
}

// Uploaded is used to set the status of a customer record in the database to "uploaded".
func (c *customer) Uploaded() error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("while starting update: %s", err)
	}

    // If returning an error Rollback the transaction, otherwise Commit it.
    defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			err = tx.Rollback()
		}
	}()

	_, err = tx.Exec(updateUploaded, string(c.Email))
	if err != nil {
		return fmt.Errorf("while updating: %s", err)
	}

	return nil
}
