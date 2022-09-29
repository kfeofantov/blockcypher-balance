FROM golang:1.19 as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /crypto-balance *.go

FROM golang:1.19

WORKDIR /app

COPY --from=builder /crypto-balance /crypto-balance
COPY --from=builder /app/templates /app/templates

ENV HTTP_ADDR="0.0.0.0:9091"

CMD /crypto-balance