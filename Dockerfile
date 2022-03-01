FROM golang:alpine

RUN apk add build-base
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY controller/*.go ./controller/
COPY database/*.go ./database/
COPY model/*.go ./model/
COPY routes/*.go ./routes/
COPY utils/*.go ./utils/

RUN go build -o /backend-service

EXPOSE 8000

CMD [ "/backend-service" ]