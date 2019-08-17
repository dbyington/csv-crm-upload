# CSV to CRM uploader

A demonstration of reading customer data from a CSV file, saving to a Postgresql database, and uploading (POST) that to a CRM service provider.

### Rules:
1. The CSV reading and CRM upload with be done by separate services
1. The CSV file will be of unknown length and must not be loaded, entirely into memory
1. Each customer's data must only be uploaded to the CRM service once
1. All service code will be written in Go using the standard Go library

### Additional notes:
- Steps need to be taken to handle failures in the CRM service provider
- Both CSV reading and CRM upload binaries will run on the same machine
