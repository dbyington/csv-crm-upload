package csvreader

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dbyington/csv-crm-upload/database"
	databaseMock "github.com/dbyington/csv-crm-upload/database/mock"
	"github.com/dbyington/csv-crm-upload/signal/sender"
)

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

var (
	errTest = errors.New("test error")

	dbMock *sql.DB
	mockDB sqlmock.Sqlmock
)

const (
	csvHeaderRow = `id,first_name,last_name,email,phone`
	goodCSV      = `1,jon,doe,jon.doe@mail.com,+1 212 555 1234`
	badCSV       = `foo,,,,,`
)

var _ = Describe("Csv", func() {
	var (
		csvString string
		err       error
		hasHeader bool
		r         *reader
		mockCtrl  *gomock.Controller

		rpcBuffer io.ReadWriteCloser
		rpcClient *rpc.Client
		rpcSender *sender.Sender

		mockCustomers *databaseMock.MockCustomers
	)

	BeforeEach(func() {
		dbMock, mockDB, err = sqlmock.New()
		mockCtrl = gomock.NewController(GinkgoT())
		mockCustomers = databaseMock.NewMockCustomers(mockCtrl)
		rpcBuffer = &buffer{}
		rpcClient = rpc.NewClient(rpcBuffer)
		rpcSender = sender.NewSender(rpcClient)
	})

	Context("NewReader", func() {
		var testR *reader
		BeforeEach(func() {
			csvString = csvHeaderRow + "\n" + goodCSV
			hasHeader = true
			r = &reader{
				csv.NewReader(strings.NewReader(csvString)),
				database.NewCustomerDB(dbMock),
				!hasHeader,
				5,
				rpcSender,
			}
			testR = NewReader(database.NewCustomerDB(dbMock),
				strings.NewReader(csvString),
				rpcClient,
				hasHeader,
				5)
		})
		It("should return a reader", func() {
			Expect(testR).ToNot(BeNil())
			Expect(testR).To(BeEquivalentTo(r))
		})
	})
	Context("dropHeaderRow", func() {
		var row []string
		var headerErr error
		Context("with a good CSV header", func() {
			JustBeforeEach(func() {
				headerErr = r.dropHeaderRow()
				row, err = r.Read()
			})

			BeforeEach(func() {
				csvString = csvHeaderRow + "\n" + goodCSV
				hasHeader = true
			})

			It("should swallow the row", func() {
				Expect(headerErr).ToNot(HaveOccurred())
				Expect(row[0]).To(Equal("1"))
			})
		})

		Context("with nothing to read", func() {
			JustBeforeEach(func() {
				headerErr = r.dropHeaderRow()
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
					rpcSender,
				}
			})

			It("should report a parse error", func() {
				Expect(headerErr).To(MatchError(io.EOF))
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
					rpcSender,
				}
			})

			It("should append to the customers", func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()

				err = r.readCustomers()
				Expect(err).To(MatchError(io.EOF))
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
					rpcSender,
				}
			})

			It("should insert and continue", func() {
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				mockDB.ExpectBegin()
				mockDB.ExpectExec("INSERT").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()
				err = r.readCustomers()
				Expect(err).To(MatchError(io.EOF))
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
					rpcSender,
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
					rpcSender,
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
					rpcSender,
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
