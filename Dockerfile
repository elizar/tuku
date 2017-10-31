FROM golang:latest

EXPOSE 8082

ENV APP_DIR $GOPATH/src/github.com/elizar/tuku

RUN mkdir -p $APP_DIR

COPY . $APP_DIR

WORKDIR $APP_DIR

# Lint, test and buidl
RUN go get -u github.com/kisielk/errcheck
RUN go get -u github.com/golang/lint/golint
RUN go get -u ./...
RUN go vet ./... && errcheck ./... && golint -set_exit_status ./...
RUN go test -v ./...
RUN go build .

ENTRYPOINT ./tuku -port ${PORT} -file ${FILE}
