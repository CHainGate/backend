FROM golang:alpine

RUN apk add build-base
WORKDIR /app

RUN apk update && apk add bash

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd/*.go ./cmd/
COPY internal/ ./internal/
COPY swaggerui/ ./swaggerui/
COPY pkg/ ./pkg/
COPY websocket/ ./websocket/
COPY wait-for-it.sh ./
COPY .openapi-generator-ignore ./

# maybe there is a better way to use openapi-generator-cli
RUN apk add --update nodejs npm
RUN apk add openjdk11
RUN npm install @openapitools/openapi-generator-cli -g
RUN npx @openapitools/openapi-generator-cli generate -i ./swaggerui/config/openapi.yaml -g go-server -o ./ --additional-properties=sourceFolder=configApi,packageName=configApi
RUN npx @openapitools/openapi-generator-cli generate -i ./swaggerui/public/openapi.yaml -g go-server -o ./ --additional-properties=sourceFolder=publicApi,packageName=publicApi
RUN npx @openapitools/openapi-generator-cli generate -i ./swaggerui/internal/openapi.yaml -g go-server -o ./ --additional-properties=sourceFolder=internalApi,packageName=internalApi
RUN npx @openapitools/openapi-generator-cli generate -i https://raw.githubusercontent.com/CHainGate/proxy-service/main/swaggerui/openapi.yaml -g go -o ./proxyClientApi --ignore-file-override=.openapi-generator-ignore --additional-properties=sourceFolder=proxyClientApi,packageName=proxyClientApi
RUN npx @openapitools/openapi-generator-cli generate -i https://raw.githubusercontent.com/CHainGate/ethereum-service/main/swaggerui/openapi.yaml -g go -o ./ethClientApi --ignore-file-override=.openapi-generator-ignore --additional-properties=sourceFolder=ethClientApi,packageName=ethClientApi
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN goimports -w .

RUN ["chmod", "+x", "wait-for-it.sh"]

RUN go build -o /backend-service ./cmd/main.go

EXPOSE 8000

CMD [ "/backend-service" ]