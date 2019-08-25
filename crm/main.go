package main

import (
    "database/sql"
    "fmt"
    "github.com/dbyington/csv-crm-upload/crm/upload"
    "github.com/dbyington/csv-crm-upload/database"
    "log"
    "os"
)

const crmAPI = "/customers"

func main() {
    dbUser := os.Getenv("POSTGRES_CSV_USER")
    dbPassword := os.Getenv("POSTGRES_CSV_PASSWORD")
    dbHost := os.Getenv("POSTGRES_HOST")
    dbName := os.Getenv("POSTGRES_DATABASE")

    connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)
    d, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("while opening database: %s", err)
    }
    db := database.NewCustomerDB(d)
    log.Print("database open")
    defer db.Close()

    listenerAddr := os.Getenv("CRM_LISTENER_ADDR")
    crmServerAddr := os.Getenv("CRM_SERVER_ADDR")

    uploader := upload.NewUploader(listenerAddr, crmServerAddr, crmAPI, db)
    uploader.Start()
}
