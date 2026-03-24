module github.com/LouYuanbo1/go-webservice/imageutil

go 1.25.4

replace github.com/LouYuanbo1/go-webservice/errorx => ../errorx

require (
	github.com/LouYuanbo1/go-webservice/errorx v0.0.0-00010101000000-000000000000
	github.com/disintegration/imaging v1.6.2
)

require golang.org/x/image v0.25.0 // indirect
