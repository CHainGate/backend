# backend


swagger urls: \
config: http://localhost:8000/api/config/swaggerui/ \
public: http://localhost:8000/api/public/swaggerui/ \
internal: http://localhost:8000/api/internal/swaggerui/ \

openapi gen:
 ```
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/swaggerui/config/openapi.yaml -g go-server -o /local/ --additional-properties=sourceFolder=configApi,packageName=configApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/swaggerui/public/openapi.yaml -g go-server -o /local/ --additional-properties=sourceFolder=publicApi,packageName=publicApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/swaggerui/internal/openapi.yaml -g go-server -o /local/ --additional-properties=sourceFolder=internalApi,packageName=internalApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i https://raw.githubusercontent.com/CHainGate/proxy-service/main/swaggerui/openapi.yaml -g go -o /local/proxyClientApi --ignore-file-override=/local/.openapi-generator-ignore --additional-properties=sourceFolder=proxyClientApi,packageName=proxyClientApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i https://raw.githubusercontent.com/CHainGate/ethereum-service/main/swaggerui/openapi.yaml -g go -o /local/ethClientApi --ignore-file-override=/local/.openapi-generator-ignore --additional-properties=sourceFolder=ethClientApi,packageName=ethClientApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i https://raw.githubusercontent.com/CHainGate/bitcoin-service/main/swaggerui/openapi.yaml -g go -o /local/btcClientApi --ignore-file-override=/local/.openapi-generator-ignore --additional-properties=sourceFolder=btcClientApi,packageName=btcClientApi
goimports -w .
 ```
 
 Code Coverage 80%
 
