package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/template"

	// This external lib is required for postgres.
	_ "github.com/lib/pq"
)

const (
	insertTemplate      = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORD(null::customers, '{{.Single}}');`
	insertSetTemplate   = `INSERT INTO customers SELECT * FROM JSON_POPULATE_RECORDSET(null::customers, {{.Set}});`
	selectUploadedFalse = `SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false;`
	updateUploaded      = `UPDATE customers SET uploaded = true WHERE email = $1;`
)

var (
	insertCustomerTemplate    *template.Template
	insertCustomerSetTemplate *template.Template
)

type Customer struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type templateFields struct {
	Single string
	Set    string
}

func init() {
	insertCustomerTemplate = template.Must(template.New("insertCustomer").Parse(insertTemplate))
	insertCustomerSetTemplate = template.Must(template.New("insertCustomerSet").Parse(insertSetTemplate))
}

// DB sets up the DB interface to our postgres database and handles the connection.
type dbase struct {
	sqlDB *sql.DB
}

type DB interface {
	InsertCustomer([]byte) error
	InsertCustomerSet([]byte) error
	SelectCustomersForUpload() ([]byte, error)
}

// New DB returns an instance of this DB or an error on failure to open a connection to the database.
func NewDB(user, password, host, database string) (*dbase, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, database)
	d, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("while opening database: %s", err)
	}

	return &dbase{sqlDB: d}, nil
}

// InsertCustomer takes a single Customer JSON object in byte slice form to insert into the database.
func (d *dbase) InsertCustomer(customer []byte) error {
	s := &templateFields{Single: string(customer)}
	return d.insert(s)
}

// InsertCustomerSet takes a byte slice containing a JSON formatted array of Customer JSON objects to do a bulk insert.
func (d *dbase) InsertCustomerSet(customerSet []byte) error {
	s := &templateFields{Set: string(customerSet)}
	return d.insert(s)
}

func (d *dbase) insert(s *templateFields) error {
	b := new(bytes.Buffer)
	err := insertCustomerTemplate.Execute(b, s)
	if err != nil {
		return fmt.Errorf("customer data: %s", err)
	}

	tx, err := d.sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("while creating transaction: %s", err)
	}
	defer func() {
        switch err {
        case nil:
            err = tx.Commit()
        default:
            err = tx.Rollback()
        }
    }()

	_, err = tx.Exec(b.String())
	if err != nil {
		return fmt.Errorf("executing query: %s", err)
	}

	return nil
}

// SelectCustomersForUpload returns a slice of *Customer structs suitable to be marshaled and uploaded to CRM.
func (d *dbase) SelectCustomersForUpload() ([]*Customer, error) {
	var customers []*Customer

	rows, err := d.sqlDB.Query(selectUploadedFalse)
	if err != nil {
		return nil, fmt.Errorf("while selecting rows: %s", err)
	}

	for rows.Next() {
		c := new(Customer)
		err := rows.Scan(&c.Id, &c.FirstName, &c.LastName, &c.Email, &c.Phone)
		if err != nil {
			return nil, fmt.Errorf("while scanning rows: %s", err)
		}

		customers = append(customers, c)
	}
	return customers, nil
}

// UpdateUploaded is used to set the status of a customer record in the database to "uploaded".
func (d *dbase) UpdateUploaded(email []byte) error {
    tx, err := d.sqlDB.Begin()
    if err != nil {
        return fmt.Errorf("while starting update: %s", err)
    }
    defer func() {
        switch err {
        case nil:
            err = tx.Commit()
        default:
            err = tx.Rollback()
        }
    }()

    _, err = tx.Exec(updateUploaded, string(email))
    if err != nil {
        return fmt.Errorf("while updating: %s", err)
    }

    return nil
}


