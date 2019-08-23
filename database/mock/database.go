package mock_database

import "database/sql"

type customer struct {
   Id        int64  `json:"id"`
   FirstName string `json:"first_name"`
   LastName  string `json:"last_name"`
   Email     string `json:"email"`
   Phone     string `json:"phone"`
}

type customers []*customer

type cdb struct {
    *sql.DB
}

type CustomerDB interface {
    NewCustomer(int64, string, string, string, string) *customer
    SelectCustomersForUpload() (*customers, error)
}

func NewCDB(d *sql.DB) *cdb {
    return &cdb{d}
}

//
//func NewCustomer(id int64, first, last, email, phone string) *customer {
//    return &customer{id, first, last, email, phone}
//}
