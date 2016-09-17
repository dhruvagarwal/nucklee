##NUCKLEE
A mocking service for APIs. You provide a requests file in the format of `httpie` and this service will mock those requests(url, method) and return the stored response(in the http requests file).

How to run
===========

    cd $GOPATH/src/github.com/dhruvagarwal/nucklee
    go run nucklee.go -path . -port 12345

Future Scope
* Allow request headers to decide response
* Allow multiple responses per url and best matching 
request or randomization as picking strategies
