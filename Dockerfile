FROM golang:alpine

RUN apk add build-base
WORKDIR /app

RUN apk update && apk add bash

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY controller/*.go ./controller/
COPY database/*.go ./database/
COPY model/*.go ./model/
COPY routes/*.go ./routes/
COPY utils/*.go ./utils/
COPY wait-for-it.sh ./

RUN ["chmod", "+x", "wait-for-it.sh"]
RUN go build -o /backend-service

EXPOSE 8000

CMD [ "/backend-service" ]