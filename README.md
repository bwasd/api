# API
Simple proof of concept web-service / API in Go.

<!-- BEGIN generated code . -->
1.	[ API](#api)
	1.	[ Requirements](#requirements)
	2.	[ Installation](#installation)
		1.	[ Database setup](#database-setup)
		2.	[ Start the server](#start-the-server)
	3.	[ Usage](#usage)
		1.	[ Error model](#error-model)
		2.	[ Create](#create)
		3.	[ Lookup](#lookup)
		4.	[ List](#list)
		5.	[ Health](#health)
<!-- END generated code . -->

## Requirements
```
go: >= 1.15
postgresql >= 11.x
```

## Installation
```sh
apt-get install postgresql postgresql-client
sudo -u postgres psql postgres
```

After dropping into the postgres shell change the password
```sh
#\password
Enter new password:
123456
\q
```

### Database setup

```
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=123456
export DB_NAME=test
```

Run `go test` in the `storage` package to initialize the database schema
(see `storage_test.go`).

### Start the server
```sh
make run
```
or equivalently 
```sh
go run cmd/api/api.go
```

## Usage

```sql
psql -U postgres -d test -a -f storage/00.sql 
```

### Create
Create a product.

Example request

```sh
curl --header "Content-Type: application/json" \
	--request POST \
	--data '{"sku":"9a982384y932", "attrs":{"mass":"undefined"}}' \
    http://localhost:8080/product/create
```

### Lookup
Lookup a product by SKU.

Example request
```sh
curl "http://localhost:8080/product/get?sku=9a982384y932" | jq
```

### List 
List products (paginated).

Example request
```sh
curl "http://localhost:8080/product/list?lo=0&hi=10" | jq
```

### Delete
Delete a product

Example request
```sh
curl "http://localhost:8080/product/delete?sku=9a982384y932" | jq

### Health
The healthz endpoint returns the server's health status; when the server
is started if backing systems are unreachable, all other endpoints will
respond with `HTTP 503 ServiceUnavailable`

Example request
```sh
curl http://localhost:8080/healthz
```
