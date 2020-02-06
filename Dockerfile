FROM golang:1.14rc1-buster as builder

WORKDIR /app

COPY . .

RUN go build

FROM ubuntu:18.04

WORKDIR /app

RUN apt-get update && apt-get install ca-certificates -y

COPY --from=builder /app/pincode_api .

EXPOSE 8080

CMD [ "/app/pincode_api" ]