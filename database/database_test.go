package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	database            = `crm_test`
	host                = `localhost`
	goodCustomerJSON    = `{ "id": 1, "first_name": "jon", "last_name": "doe", "email": "jdoe@mail.com", "phone": "+1 212 555 1234"}`
	goodCustomerSetJSON = `[ { "id": 1, "first_name": "jon", "last_name": "doe", "email": "jon.doe@mail.com", "phone": "+1 212 555 1234"},
{ "id": 2, "first_name": "jane", "last_name": "doe", "email": "jane.doe@mail.com", "phone": "+1 212 555 4321"},
{ "id": 3, "first_name": "steve", "last_name": "stevenson", "email": "steves@mail.com", "phone": "+1 503 555 5522"} ]`
)

var (
	password string
	user     string

	err     error
	errTest = errors.New("test error")

	dbMock *sql.DB
	mockDB sqlmock.Sqlmock
)

func init() {
	password = os.Getenv("POSTGRES_CSV_PASSWORD")
	user = os.Getenv("POSTGRES_CSV_USER")
}

var _ = Describe("Database", func() {
	var (
		d *dbase
	)

	BeforeEach(func() {
		dbMock, mockDB, err = sqlmock.New()
	})

	Context("NewDB", func() {

		Context("with valid database setup", func() {
			BeforeEach(func() {
				d, err = NewDB(user, password, host, database)
			})

			AfterEach(func() {
				d = nil
			})

			It("should return a database", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(d.sqlDB).ToNot(BeNil())
			})
		})
	})

	Context(".insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
		)

		BeforeEach(func() {
			d = &dbase{sqlDB: dbMock}
		})

		AfterEach(func() {
			d = nil
		})

		Context("with a single good customer", func() {
			BeforeEach(func() {
				s := &templateFields{Single: goodCustomerJSON}

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = d.insert(s)
			})
			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with a good customer set", func() {
			BeforeEach(func() {
				s := &templateFields{Set: goodCustomerSetJSON}

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(3, 3))
				mockDB.ExpectCommit()
				err = d.insert(s)
			})
			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with transaction begin failure", func() {
			BeforeEach(func() {
				s := &templateFields{Single: goodCustomerJSON}

				mockDB.ExpectBegin().WillReturnError(errTest)
				err = d.insert(s)
			})
			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while creating transaction: %s", errTest)))
			})
		})

		Context("with a bad customer json string", func() {
			BeforeEach(func() {
				s := &templateFields{Single: "bad customer"}

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnError(errTest)
				mockDB.ExpectRollback()
				err = d.insert(s)
			})
			It("should return an error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("executing query: %s", errTest)))
			})
		})
	})

	Context(".InsertCustomer", func() {
		var (
			expectedInsert = `INSERT INTO customers`
		)

		BeforeEach(func() {
			d = &dbase{sqlDB: dbMock}
		})

		AfterEach(func() {
			d = nil
		})

		Context("with a single good customer", func() {
			BeforeEach(func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = d.InsertCustomer([]byte(goodCustomerJSON))
			})
			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})
	})

	Context(".InsertCustomerSet", func() {
		var (
			expectedInsert = `INSERT INTO customers`
		)

		BeforeEach(func() {
			d = &dbase{sqlDB: dbMock}
		})

		AfterEach(func() {
			d = nil
		})

		Context("with a good customer set", func() {
			BeforeEach(func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = d.InsertCustomerSet([]byte(goodCustomerSetJSON))
			})
			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})
	})

	Context(".SelectCustomersForUpload", func() {
		var expectedRows *sqlmock.Rows
		var rowsReturned []*Customer

		BeforeEach(func() {
			d = &dbase{sqlDB: dbMock}
		})

		AfterEach(func() {
			d = nil
		})

		Context("with a successful select", func() {
			BeforeEach(func() {
				expectedRows = sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone"}).
					AddRow(1, "jon", "doe", "jon.doe@mail.com", "+1 212 555 1234").
					AddRow(2, "jane", "doe", "jane.doe@mail.com", "+1 212 555 4321").
					AddRow(3, "steve", "stevenson", "steves@mail.com", "+1 503 555 5522")
				mockDB.ExpectQuery("SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false").WillReturnRows(expectedRows)
				rowsReturned, err = d.SelectCustomersForUpload()
			})

			It("should return customers needing to be uploaded", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).ToNot(HaveOccurred())
				Expect(rowsReturned).ToNot(BeNil())
				Expect(len(rowsReturned)).To(Equal(3))
			})
		})

		Context("when an error occurs selecting rows", func() {
			BeforeEach(func() {
				mockDB.ExpectQuery("SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false").WillReturnError(errTest)
				rowsReturned, err = d.SelectCustomersForUpload()
			})

			It("should return a select error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while selecting rows: %s", errTest)))
				Expect(rowsReturned).To(BeNil())
			})
		})

		Context("when an error occurs scanning rows", func() {
			BeforeEach(func() {
				expectedRows = sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone"}).
					AddRow(1, "jon", "doe", "jdoe@mail.com", "+1 212 555 1234").
					RowError(0, errTest)
				mockDB.ExpectQuery("SELECT").WillReturnRows(expectedRows)
				rowsReturned, err = d.SelectCustomersForUpload()
			})

			It("should return a scan error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				//Expect(err).To(MatchError(fmt.Errorf("while selecting rows: %s", errTest)))
				Expect(rowsReturned).To(BeNil())
			})
		})
	})

	Context(".UpdateUploaded", func() {
		BeforeEach(func() {
			d = &dbase{sqlDB: dbMock}
		})

		AfterEach(func() {
			d = nil
		})

		Context("with a successful update", func() {
		    BeforeEach(func() {
		        mockDB.ExpectBegin()
		        mockDB.ExpectExec("UPDATE customers").WillReturnResult(sqlmock.NewResult(1, 1))
		        mockDB.ExpectCommit()
		        err = d.UpdateUploaded([]byte("jdoe@mail.com"))
            })

		    It("should not return an error", func() {
		        Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
		        Expect(err).ToNot(HaveOccurred())
            })
        })

		Context("with a failed update", func() {
		    BeforeEach(func() {
		        mockDB.ExpectBegin()
		        mockDB.ExpectExec("UPDATE customers").WillReturnError(errTest)
		        mockDB.ExpectRollback()
		        err = d.UpdateUploaded([]byte("jdoe@mail.com"))
            })

		    It("should return an error", func() {
		        Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
		        Expect(err).To(MatchError(fmt.Errorf("while updating: %s", errTest)))
            })
        })
	})
})
