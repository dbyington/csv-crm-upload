package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/dbyington/csv-crm-upload/cmd/csvreader"
	"github.com/dbyington/csv-crm-upload/database"
)

func main() {
	var (
		csvFileName     string
		csvNoHeaderRow  bool
		dbUser          string
		dbPassword      string
		dbHost          string
		dbName          string
		bufferSize      int
		listenerAddress string
		listenerNet     string
	)
	flag.StringVar(&csvFileName, "filename", os.Getenv("CSV_FILE"), "Path to the CSV file containing the customer records to upload.")
	flag.BoolVar(&csvNoHeaderRow, "noheader", false, "Used if the CSV file does not contain a header row.")
	flag.IntVar(&bufferSize, "buffer", 5, "Number of lines to read in before writing to the database and signalling the CRM upload worker")
	flag.StringVar(&dbUser, "username", os.Getenv("POSTGRES_CSV_USER"), "Username used to connect to the postgres database.")
	flag.StringVar(&dbPassword, "password", os.Getenv("POSTGRES_CSV_PASSWORD"), "Password used to connect to the postgres database.")
	flag.StringVar(&dbHost, "dbhost", os.Getenv("POSTGRES_HOST"), "Hostname used to connect to the postgres database.")
	flag.StringVar(&dbName, "database", os.Getenv("POSTGRES_DATABASE"), "Username used to connect to the postgres database.")
	flag.StringVar(&listenerAddress, "rpcaddr", "localhost:9876", "Hostname used to connect to the signal listener.")
	flag.StringVar(&listenerNet, "rpcnetwork", "tcp", "Network used to connect to the signal listener")
	flag.Parse()

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)
	d, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("while opening database: %s", err)
	}
	db := database.NewCustomerDB(d)
	log.Print("database open")
	defer db.Close()

	file, err := os.Open(csvFileName)
	if err != nil {
		log.Fatalf("while opening CSV file: %s", err)
	}
	log.Print("csv file open")
	defer file.Close()

	rpcClient, err := rpcDial(listenerNet, listenerAddress)
	if err != nil {
		log.Fatalf("while dialing server: %s", err)
	}
	defer rpcClient.Close()

	reader := csvreader.NewReader(db, file, rpcClient, csvNoHeaderRow, bufferSize)
	log.Println("starting...")
	if err := reader.Run(); err != nil {
		log.Printf("error reading: %s", err)
	}
	log.Println("done.")
}

func rpcDial(n, a string) (*rpc.Client, error) {
	// All of this is kind of funky but it's setting up a timer to force a timeout if the rpc dial takes too long.

	timeout := 5 // A 5 second timeout seems reasonable.
	var c *rpc.Client
	var err error
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	// Setup a channel to signal on when the rpc dial returns.
	clientReady := make(chan struct{})
	go func() {
		c, err = rpc.DialHTTP(n, a)
		clientReady <- struct{}{}
	}()

	// Wait here for either the rpc dial to return and cancel the timer or the timer finishes.
	select {
	case <-clientReady:
		timer.Stop()
	case <-timer.C:
		log.Fatal("timeout waiting for server")
	}

	// Now handle the error set by rpc.DialHTTP()
	if err != nil {
		return nil, err
	}

	return c, nil
}
