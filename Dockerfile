FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /crypto-balance main.go

FROM golang:alpine

COPY --from=builder /crypto-balance /crypto-balance

ENV HTTP_ADDR="0.0.0.0:9091"

CMD /crypto-balance