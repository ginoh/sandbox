FROM golang

EXPOSE 8080
WORKDIR /go/src/app

COPY go.mod .
RUN go mod download -x

COPY . .
RUN go install -v

CMD ["sample-server"]
