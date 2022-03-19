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
COPY models/*.go ./models/
COPY openApiSpecifications/* ./openApiSpecifications/
COPY routes/*.go ./routes/
COPY service/ ./service/
COPY swaggerui/ ./swaggerui/
COPY utils/*.go ./utils/
COPY wait-for-it.sh ./
COPY .openapi-generator-ignore ./

# maybe there is a better way to use openapi-generator-cli
RUN apk add --update nodejs npm
RUN apk add openjdk11
RUN npm install @openapitools/openapi-generator-cli -g
RUN npx @openapitools/openapi-generator-cli generate -i ./openApiSpecifications/config.yaml -g go-server -o ./ --additional-properties=sourceFolder=configApi,packageName=configApi
RUN npx @openapitools/openapi-generator-cli generate -i ./openApiSpecifications/public.yaml -g go-server -o ./ --additional-properties=sourceFolder=publicApi,packageName=publicApi
RUN npx @openapitools/openapi-generator-cli generate -i ./openApiSpecifications/internal.yaml -g go-server -o ./ --additional-properties=sourceFolder=internalApi,packageName=internalApi
RUN npx @openapitools/openapi-generator-cli generate -i ./openApiSpecifications/proxy.yaml -g go -o ./proxyClientApi --ignore-file-override=.openapi-generator-ignore --additional-properties=sourceFolder=proxyClientApi,packageName=proxyClientApi
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN goimports -w .

RUN ["chmod", "+x", "wait-for-it.sh"]

RUN go build -o /backend-service

EXPOSE 8000

CMD [ "/backend-service" ]