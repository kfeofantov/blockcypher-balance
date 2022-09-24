FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /crypto-balance *.go

FROM golang:alpine

WORKDIR /app

COPY --from=builder /crypto-balance /crypto-balance
COPY --from=builder /app/templates /app/templates

ENV HTTP_ADDR="0.0.0.0:9091"

CMD /crypto-balance