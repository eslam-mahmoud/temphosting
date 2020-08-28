FROM golang:1.14

WORKDIR /go/src/app

COPY . .

ENV GO111MODULE=on

# get module
# go get -u ./... from your module root upgrades all the direct and indirect dependencies of your module, and now excludes test dependencies.
# go get -u -t ./... is similar, but also upgrades test dependencies
RUN go get -u -t ./...

# should build and run executable
RUN go build ./cmd/webserver/

EXPOSE 8080

CMD ["./webserver"]