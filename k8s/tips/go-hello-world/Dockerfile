FROM golang AS builder

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o hello main.go

FROM scratch

COPY --from=builder /go/src/app/hello hello

CMD ["./hello"]
