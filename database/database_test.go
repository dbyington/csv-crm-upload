package database

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	database            = `crm`
	host                = `localhost`
	goodCustomerJSON    = `{ "id": 1, "first_name": "jon", "last_name": "doe", "email": "jdoe@mail.com", "phone": "+1 212 555 1234"}`
	goodCustomerSetJSON = `[ { "id": 1, "first_name": "jon", "last_name": "doe", "email": "jon.doe@mail.com", "phone": "+1 212 555 1234"},
{ "id": 2, "first_name": "jane", "last_name": "doe", "email": "jane.doe@mail.com", "phone": "+1 212 555 4321"},
{ "id": 3, "first_name": "steve", "last_name": "stevenson", "email": "steves@mail.com", "phone": "+1 503 555 5522"} ]`
)

var (
	err     error
	errTest = errors.New("test error")

	dbMock *sql.DB
	mockDB sqlmock.Sqlmock
)

var _ = Describe("Database", func() {
	var (
		expectedCustomer1 = &Customer{
			Id:        1,
			FirstName: "jon",
			LastName:  "doe",
			Email:     "jon.doe@mail.com",
			Phone:     "+1 212 555 1234",
		}
		expectedCustomer2 = &Customer{
			Id:        2,
			FirstName: "jane",
			LastName:  "doe",
			Email:     "jane.doe@mail.com",
			Phone:     "+1 212 555 4321",
		}
	)

	BeforeEach(func() {
		dbMock, mockDB, err = sqlmock.New()
	})

	Context("Open", func() {
		BeforeEach(func() {
			db = nil
		})

		Context("when called", func() {
			BeforeEach(func() {
				Open("postgres", "postgres", host, database)
			})

			It("should open the database", func() {
				Expect(db).ToNot(BeNil())
			})
		})
	})

	Context("NewCustomer", func() {
		var testCustomer *Customer

		Context("when called", func() {
			BeforeEach(func() {
				testCustomer = NewCustomer(1, "jon", "doe", "jon.doe@mail.com", "+1 212 555 1234")
			})

			It("should return a customer", func() {
				Expect(testCustomer).To(BeEquivalentTo(expectedCustomer1))
			})
		})
	})

	Context("NewCustomers", func() {
		var testCustomers *Customers
		expectedCustomers := new(Customers)

		AfterEach(func() {
			expectedCustomers = nil
		})

		Context("when called with customers", func() {
			BeforeEach(func() {
				*expectedCustomers = append(*expectedCustomers, expectedCustomer1, expectedCustomer2)
				testCustomers = NewCustomers(expectedCustomer1, expectedCustomer2)
			})

			It("should return a set of customers", func() {
				Expect(testCustomers).To(BeEquivalentTo(expectedCustomers))
			})
		})

		Context("when called without customers", func() {
			BeforeEach(func() {
				testCustomers = NewCustomers()
			})

			It("should return an empty set of customers", func() {
				Expect(testCustomers).To(BeEquivalentTo(&Customers{}))
			})
		})
	})

	Context("insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
		)

		BeforeEach(func() {
			db = dbMock
		})

		AfterEach(func() {
			db = nil
		})

		Context("with a single good customer", func() {
			BeforeEach(func() {
				s := &templateFields{JSON: goodCustomerJSON}
				b := new(bytes.Buffer)
				err = insertCustomerTemplate.Execute(b, s)
				Expect(err).ToNot(HaveOccurred())

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = insert(b.Bytes())
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with a good customer set", func() {
			BeforeEach(func() {
				s := &templateFields{JSON: goodCustomerSetJSON}
				b := new(bytes.Buffer)
				err = insertCustomerSetTemplate.Execute(b, s)
				Expect(err).ToNot(HaveOccurred())

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(3, 3))
				mockDB.ExpectCommit()
				err = insert(b.Bytes())
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with transaction begin failure", func() {
			BeforeEach(func() {
				s := &templateFields{JSON: goodCustomerJSON}
				b := new(bytes.Buffer)
				err = insertCustomerTemplate.Execute(b, s)
				Expect(err).ToNot(HaveOccurred())

				mockDB.ExpectBegin().WillReturnError(errTest)
				err = insert(b.Bytes())
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while creating transaction: %s", errTest)))
			})
		})

		Context("with a bad customer json string", func() {
			BeforeEach(func() {
				s := &templateFields{JSON: "bad customer"}
				b := new(bytes.Buffer)
				err = insertCustomerTemplate.Execute(b, s)
				Expect(err).ToNot(HaveOccurred())

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnError(errTest)
				mockDB.ExpectRollback()
				err = insert(b.Bytes())
			})

			It("should return an error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("executing query: %s", errTest)))
			})
		})
	})

	Context("Customer.Insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
			testCustomer   *Customer
		)

		BeforeEach(func() {
			db = dbMock
			testCustomer = &Customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
			}
		})

		AfterEach(func() {
			db = nil
		})

		Context("with a single good customer", func() {
			BeforeEach(func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = testCustomer.Insert()
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})
	})

	Context("Customers.Insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
			testCustomers  = &Customers{}
		)

		BeforeEach(func() {
			db = dbMock

			*testCustomers = append(*testCustomers, &Customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
			})
		})

		AfterEach(func() {
			db = nil
		})

		Context("with a good customer set", func() {
			BeforeEach(func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = testCustomers.Insert()
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})
	})

	Context(".Append", func() {
        var testCustomers = &Customers{}
        var expectedCustomers = &Customers{}

        BeforeEach(func() {
            *expectedCustomers = append(*expectedCustomers, expectedCustomer1, expectedCustomer2)
            *testCustomers = append(*testCustomers, expectedCustomer1)

            testCustomers.Append(expectedCustomer2)
        })

        It("should add the supplied customer", func() {
            Expect(testCustomers).To(BeEquivalentTo(expectedCustomers))
        })
    })

	Context(".Count", func() {
        var testCustomers = &Customers{}

        Context("with no customers", func() {
            It("should return 0", func() {
                Expect(testCustomers.Count()).To(Equal(0))
            })
        })

        Context("with 2 customers", func() {
            BeforeEach(func() {
                *testCustomers = append(*testCustomers, expectedCustomer1, expectedCustomer2)
            })

            It("should return 2", func() {
                Expect(testCustomers.Count()).To(Equal(2))
            })
        })
    })

	Context("SelectCustomersForUpload", func() {
		var expectedRows *sqlmock.Rows
		var rowsReturned *Customers

		BeforeEach(func() {
			db = dbMock
		})

		AfterEach(func() {
			db = nil
		})

		Context("with a successful select", func() {
			BeforeEach(func() {
				expectedRows = sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone"}).
					AddRow(1, "jon", "doe", "jon.doe@mail.com", "+1 212 555 1234").
					AddRow(2, "jane", "doe", "jane.doe@mail.com", "+1 212 555 4321").
					AddRow(3, "steve", "stevenson", "steves@mail.com", "+1 503 555 5522")
				mockDB.ExpectQuery("SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false").WillReturnRows(expectedRows)
				rowsReturned, err = SelectCustomersForUpload()
			})

			It("should return customers needing to be uploaded", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).ToNot(HaveOccurred())
				Expect(rowsReturned).ToNot(BeNil())
				Expect(len(*rowsReturned)).To(Equal(3))
			})
		})

		Context("when an error occurs selecting rows", func() {
			BeforeEach(func() {
				mockDB.ExpectQuery("SELECT id, first_name, last_name, email, phone FROM customers WHERE uploaded = false").WillReturnError(errTest)
				rowsReturned, err = SelectCustomersForUpload()
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
					AddRow(nil, "jon", "doe", "jdoe@mail.com", "+1 212 555 1234").
					RowError(1, errTest)
				mockDB.ExpectQuery("SELECT").WillReturnRows(expectedRows)
				rowsReturned, err = SelectCustomersForUpload()
			})

			It("should return a scan error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while scanning rows: sql: Scan error on column index 0, name \"id\": converting driver.Value type <nil> (\"<nil>\") to a int64: invalid syntax")))
				Expect(rowsReturned).To(BeNil())
			})
		})
	})

	Context(".Uploaded", func() {
		var testCustomer *Customer
		BeforeEach(func() {
			db = dbMock
			testCustomer = &Customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
			}
		})

		AfterEach(func() {
			db = nil
		})

		Context("with a successful update", func() {
			BeforeEach(func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec("UPDATE customers").WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = testCustomer.Uploaded()
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
				err = testCustomer.Uploaded()
			})

			It("should return an error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while updating: %s", errTest)))
			})
		})
	})
})
