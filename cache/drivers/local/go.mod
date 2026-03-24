module github.com/LouYuanbo1/go-webservice/cache/drivers/local

go 1.25.4

require (
	github.com/LouYuanbo1/go-webservice/cache v0.0.0-00010101000000-000000000000
	github.com/LouYuanbo1/go-webservice/errorx v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/ristretto/v2 v2.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

replace github.com/LouYuanbo1/go-webservice/errorx => ../../../errorx

replace github.com/LouYuanbo1/go-webservice/cache => ../../../cache
