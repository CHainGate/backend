# backend

openapi gen:
 ```
 docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/petstore.yaml -g go-server -o /local/ --additional-properties=sourceFolder=openapi,serverPort=8000
 ```