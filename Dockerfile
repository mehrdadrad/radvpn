FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build .

EXPOSE 8085

ENTRYPOINT ["/app/radvpn"]