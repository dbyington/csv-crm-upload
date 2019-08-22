package csvreader

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dbyington/csv-crm-upload/database"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	databaseMock "github.com/dbyington/csv-crm-upload/database/mock"
)

var (
	errTest = errors.New("test error")

	dbMock *sql.DB
	mockDB sqlmock.Sqlmock
)

const (
	csvHeaderRow    = `id,first_name,last_name,email,phone`
	goodCSV         = `1,jon,doe,jon.doe@mail.com,+1 212 555 1234`
	badCSV          = `foo,,,,,`
)

var _ = Describe("Csv", func() {
	var (
		csvString string
		err       error
		hasHeader bool
		r         *reader
		mockCtrl  *gomock.Controller

		//mockCustomer *databaseMock.MockCustomer
		mockCustomers *databaseMock.MockCustomers
	)

	BeforeEach(func() {
		dbMock, mockDB, err = sqlmock.New()
		mockCtrl = gomock.NewController(GinkgoT())
		mockCustomers = databaseMock.NewMockCustomers(mockCtrl)
	})

	Context("NewReader", func() {
		var testR *reader
		BeforeEach(func() {
			csvString = csvHeaderRow + "\n" + goodCSV
			hasHeader = true
			r = &reader{
				csv.NewReader(strings.NewReader(csvString)),
				database.NewCustomerDB(dbMock),
				hasHeader,
				5,
			}
			testR = NewReader(database.NewCustomerDB(dbMock),
				strings.NewReader(csvString),
				hasHeader,
				5)
		})
		It("should return a reader", func() {
			Expect(testR).To(BeEquivalentTo(r))
		})
	})
	Context("dropHeaderRow", func() {
		var row []string
		Context("with a good CSV header", func() {
			JustBeforeEach(func() {
				r.dropHeaderRow()
				row, err = r.Read()
			})

			BeforeEach(func() {
				csvString = csvHeaderRow + "\n" + goodCSV
				hasHeader = true
			})

			It("should swallow the row", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(row[0]).To(Equal("1"))
			})
		})

		Context("with nothing to read", func() {
			JustBeforeEach(func() {
				r.dropHeaderRow()
				_, err = r.Read()
			})

			BeforeEach(func() {
				csvString = ""
				hasHeader = true
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}
			})

			It("should report a parse error", func() {
				Expect(err).To(MatchError(io.EOF))
			})
		})
	})

	Context("readCustomers", func() {
		Context("with a good customer row", func() {
			BeforeEach(func() {
				csvString = goodCSV
				hasHeader = false
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}

			})

			It("should append to the customers", func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()

				//mockCustomers.EXPECT().Count().Return(1)
				//mockCustomers.EXPECT().Append(gomock.Any()).Times(0)
				r.readCustomers()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the buffer fills", func() {
			BeforeEach(func() {
				csvString = csvHeaderRow + "\n" + goodCSV + "\n" + goodCSV + "\n" +
					goodCSV + "\n" + goodCSV + "\n" + goodCSV + "\n" + goodCSV +
					"\n" + goodCSV
				hasHeader = true
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}
			})

			It("should insert and reset", func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				//mockCustomers.EXPECT().Append(gomock.Any()).Times(7)
				r.readCustomers()

			})
		})
	})

	Context("parseRow", func() {
		var id int64
		var first, last, email, phone string
		var err error

		Context("with a good row", func() {
			BeforeEach(func() {
				csvString = goodCSV
				hasHeader = false
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}
				id, first, last, email, phone, err = r.parseRow()
			})

			It("should return the customer record parts", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(id).To(Equal(int64(1)))
				Expect(first).To(Equal("jon"))
				Expect(last).To(Equal("doe"))
				Expect(email).To(Equal("jon.doe@mail.com"))
				Expect(phone).To(Equal("+1 212 555 1234"))
			})
		})

		Context("with a bad row id", func() {
			BeforeEach(func() {
				csvString = badCSV
				hasHeader = false
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}
				id, first, last, email, phone, err = r.parseRow()
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(fmt.Errorf("failed to parse row id: strconv.Atoi: parsing \"foo\": invalid syntax")))
			})
		})

		Context("when there are no more rows", func() {
			BeforeEach(func() {
				csvString = ""
				hasHeader = false
				r = &reader{
					csv.NewReader(strings.NewReader(csvString)),
					database.NewCustomerDB(dbMock),
					hasHeader,
					5,
				}
				id, first, last, email, phone, err = r.parseRow()
			})

			It("should return EOF error", func() {
				Expect(err).To(MatchError(io.EOF))
			})
		})
	})

	Context("insertCustomers", func() {
		BeforeEach(func() {

		})

		Context("when the customers insert succeeds", func() {
			BeforeEach(func() {
				mockCustomers.EXPECT().Insert().Times(1).Return(nil)
			})

			// TODO: When signalling is available this should be completed.
			It("should signal rows have been inserted", func() {
				Expect(nil).To(BeNil())
			})
		})

		Context("when the customers insert fails", func() {
			BeforeEach(func() {
				mockCustomers.EXPECT().Insert().Times(1).Return(errTest)
			})

			It("should get the customer list", func() {
				mockCustomers.EXPECT().List().Times(1)
			})
		})
	})
})
