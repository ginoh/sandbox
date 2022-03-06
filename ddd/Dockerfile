FROM golang

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go install ./...

CMD ["sample-api"]
