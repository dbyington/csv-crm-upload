package database

import (
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

	customerDB *cdb
	dbMock     *sql.DB
	mockDB     sqlmock.Sqlmock
)

var _ = Describe("Database", func() {
	var (
		expectedCustomer1 = &customer{
			Id:        1,
			FirstName: "jon",
			LastName:  "doe",
			Email:     "jon.doe@mail.com",
			Phone:     "+1 212 555 1234",
			Up:        false,
		}
		expectedCustomer2 = &customer{
			Id:        2,
			FirstName: "jane",
			LastName:  "doe",
			Email:     "jane.doe@mail.com",
			Phone:     "+1 212 555 4321",
			Up:        false,
		}
	)

	BeforeEach(func() {
		dbMock, mockDB, err = sqlmock.New()
		customerDB = &cdb{dbMock}
		expectedCustomer1.db = customerDB
		expectedCustomer2.db = customerDB
	})

	Context("NewCustomer", func() {
		var testCustomer *customer

		Context("when called", func() {
			BeforeEach(func() {
				testCustomer = customerDB.NewCustomer(1, "jon", "doe", "jon.doe@mail.com", "+1 212 555 1234")
			})

			It("should return a customer", func() {
				// Need to fake the timestamps for the assertion. This is the easiest way to do that.
				expectedCustomer1.Created = testCustomer.Created
				expectedCustomer1.Updated = testCustomer.Updated
				Expect(testCustomer).To(BeEquivalentTo(expectedCustomer1))
			})
		})
	})

	Context("NewCustomers", func() {
		var testCustomers *customers
		expectedCustomers := new(customers)

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
				Expect(testCustomers).To(BeEquivalentTo(&customers{}))
			})
		})
	})

	Context("insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
		)

		Context("with a single good customer", func() {
			BeforeEach(func() {

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = customerDB.insert(insertCustomer, goodCustomerJSON)
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with a good customer set", func() {
			BeforeEach(func() {

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnResult(sqlmock.NewResult(3, 3))
				mockDB.ExpectCommit()
				err = customerDB.insert(insertCustomerSet, goodCustomerSetJSON)
			})

			It("should insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
			})
		})

		Context("with transaction begin failure", func() {
			BeforeEach(func() {

				mockDB.ExpectBegin().WillReturnError(errTest)
				err = customerDB.insert(insertCustomer, goodCustomerJSON)
			})

			It("should fail to insert the customer", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while creating transaction: %s", errTest)))
			})
		})

		Context("with a bad customer json string", func() {
			BeforeEach(func() {

				mockDB.ExpectBegin()
				mockDB.ExpectExec(expectedInsert).WillReturnError(errTest)
				mockDB.ExpectRollback()
				err = customerDB.insert(insertCustomer, "bad customer")
			})

			It("should return an error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("executing query (%s): %s", insertCustomer, errTest)))
			})
		})
	})

	Context("Customer.Insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
			testCustomer   *customer
			db             *cdb
		)

		BeforeEach(func() {
			db = &cdb{dbMock}
			testCustomer = &customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
				db:        db,
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

	Context("customers.Insert", func() {
		var (
			expectedInsert = `INSERT INTO customers`
			testCustomers  = &customers{}
			db             *cdb
		)

		BeforeEach(func() {
			db = &cdb{dbMock}

			*testCustomers = append(*testCustomers, &customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
				db:        db,
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
		var testCustomers = &customers{}
		var expectedCustomers = &customers{}

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
		var testCustomers = &customers{}

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

	Context(".Count", func() {
		var testCustomers = &customers{}

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
		var (
			expectedRows *sqlmock.Rows
			rowsReturned *customers
			db           *cdb
		)

		BeforeEach(func() {
			db = &cdb{dbMock}
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
				rowsReturned, err = db.SelectCustomersForUpload()
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
				rowsReturned, err = db.SelectCustomersForUpload()
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
				rowsReturned, err = db.SelectCustomersForUpload()
			})

			It("should return a scan error", func() {
				Expect(mockDB.ExpectationsWereMet()).ToNot(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("while scanning rows: sql: Scan error on column index 0, name \"id\": converting driver.Value type <nil> (\"<nil>\") to a int64: invalid syntax")))
				Expect(rowsReturned).To(BeNil())
			})
		})
	})

	Context(".Uploaded", func() {
		var (
			db           *cdb
			testCustomer *customer
		)
		BeforeEach(func() {
			db = &cdb{dbMock}
			testCustomer = &customer{
				Id:        1,
				FirstName: "jon",
				LastName:  "doe",
				Email:     "jon.doe@mail.com",
				Phone:     "+1 212 555 1234",
				db:        db,
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
