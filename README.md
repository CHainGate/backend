# backend


swagger urls: \
config: http://localhost:8000/api/config/swaggerui/ \
public: http://localhost:8000/api/public/swaggerui/ \
internal: http://localhost:8000/api/internal/swaggerui/ \

openapi gen:
 ```
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/openApiSpecifications/config.yaml -g go-server -o /local/ --additional-properties=sourceFolder=configApi,packageName=configApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/openApiSpecifications/public.yaml -g go-server -o /local/ --additional-properties=sourceFolder=publicApi,packageName=publicApi
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/openApiSpecifications/internal.yaml -g go-server -o /local/ --additional-properties=sourceFolder=internalApi,packageName=internalApi
goimports -w .
 ```