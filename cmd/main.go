package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/dbyington/csv-crm-upload/cmd/csvreader"
	"github.com/dbyington/csv-crm-upload/database"
	"log"
	"os"
)

func main() {
	var (
		csvFileName    string
		csvNoHeaderRow bool
		dbUser         string
		dbPassword     string
		dbHost         string
		dbName         string
		lineBuffer     int
	)
	flag.StringVar(&csvFileName, "filename", os.Getenv("CSV_FILE"), "Path to the CSV file containing the customer records to upload.")
	flag.BoolVar(&csvNoHeaderRow, "noheader", false, "Used if the CSV file does not contain a header row.")
	flag.IntVar(&lineBuffer, "buffer", 5, "Number of lines to read in before writing to the database and signalling the CRM upload worker")
	flag.StringVar(&dbUser, "username", os.Getenv("POSTGRES_CSV_USER"), "Username used to connect to the postgres database.")
	flag.StringVar(&dbPassword, "password", os.Getenv("POSTGRES_CSV_PASSWORD"), "Password used to connect to the postgres database.")
	flag.StringVar(&dbHost, "host", os.Getenv("POSTGRES_HOST"), "Hostname used to connect to the postgres database.")
	flag.StringVar(&dbName, "database", os.Getenv("POSTGRES_DATABASE"), "Username used to connect to the postgres database.")
	// Parse the flags
	flag.Parse()

	// Open the database. If Open fails it will exit the program for us. No point in continuing if we cannot open the db.
	//database.Open(dbUser, dbPassword, dbHost, dbName)
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)
	d, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("while opening database: %s", err)
	}
	db := database.NewCustomerDB(d)

	file, err := os.Open(csvFileName)
	if err != nil {
		log.Fatalf("while opening CSV file: %s", err)
	}

	reader := csvreader.NewReader(db, file, csvNoHeaderRow, lineBuffer)
	reader.Run()

}
