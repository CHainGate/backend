FROM golang:alpine

RUN apk add build-base
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /backend-service

EXPOSE 8000

CMD [ "/backend-service" ]