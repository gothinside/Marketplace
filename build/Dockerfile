FROM golang:1.21

WORKDIR /go/src/shopql
COPY . /go/src/shopql
RUN go mod download
RUN make build

EXPOSE 8080/tcp

CMD [ "/go/src/shopql/bin/shopql" ]