package database

//go:generate mockgen -source=database.go -destination=mock/database_mock.go -package=mock_database

import (
    "database/sql"
    "encoding/json"
    "fmt"

    // This external lib is required for postgres.
    _ "github.com/lib/pq"
)

const (
	insertCustomer      = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORD(null::customers, $1::json);`
	insertCustomerSet   = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORDSET(null::customers, $1::json);`
	selectUploadedFalse = `SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false;`
	updateUploaded      = `UPDATE customers SET uploaded = true WHERE email = $1;`
)

type cdb struct {
	*sql.DB
}

type CustomerDB interface {
	NewCustomer(int64, string, string, string, string) *customer
	SelectCustomersForUpload() (*customers, error)
}

// Customer describes a CRM customer
type customer struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	db        *cdb   `json:"-"`
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

func NewCustomerDB(d *sql.DB) *cdb {
	return &cdb{d}
}

// NewCustomer returns a *Customer based on the supplied Customer type values.
func (db *cdb) NewCustomer(id int64, firstName, lastName, email, phone string) *customer {
	return &customer{
		id,
		firstName,
		lastName,
		email,
		phone,
		db,
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

	return c.db.insert(insertCustomer, string(jsonBytes))
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
	// Extract the db from the first customer
	var db *cdb
	if customer := c.List()[0]; customer != nil {
		db = customer.db
	} else {
		return fmt.Errorf("empty customer list")
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("while marshaling customer: %s", err)
	}

	return db.insert(insertCustomerSet, string(jsonBytes))
}

func (db *cdb) insert(query, arg string) error {
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

	_, err = tx.Exec(query, arg)
	if err != nil {
		return fmt.Errorf("executing query (%s): %s", query, err)
	}

	return nil
}

// SelectCustomersForUpload returns a *customers struct suitable to be unmarshaled and uploaded to CRM.
func (db *cdb) SelectCustomersForUpload() (*customers, error) {
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
	tx, err := c.db.Begin()
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
