module example.com/test-task

go 1.16

replace github.com/inekit/test-golang/cmd/parserProject/server => ./server

require (
	github.com/inekit/test-golang/cmd/parserProject/server
	github.com/gin-gonic/gin v1.7.3
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lib/pq v1.10.2
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)
