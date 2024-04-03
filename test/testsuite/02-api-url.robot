*** Settings ***
Resource      00-common.robot

*** Test cases ***

GET existing url by id should succeed
    Clear Expectations
    Expect Response Body    ${SCHEMA}/url.json
    Set Headers             ${admin}
    GET                     /api/url/1
    Integer                 response status                 200
    Integer                 response body id                1
    Integer                 response body db_id             1
    String                  response body db_name           db1

GET non existing url by id should fail
    Clear Expectations
    Set Headers             ${admin}
    GET                     /api/url/99
    Integer                 response status                 404

GET url by invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    GET                     /api/url/dskjhfkdsjfgh
    Integer                 response status                 400

GET all urls should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/urls.json
    GET                     /api/url
    Integer                 response status                 200
    Array                   response body                   minItems=10  maxItems=10

Create url should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/url.json
    POST                    /api/url                        {"db_id": 3, "sid": "12", "url":"postgres://user:password@127.0.0.1/db3" }
    Integer                 response status                 200
    String                  response body sid               12

Create url with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/url                        {"db_id": 3, "product_name": "vm-xl-2"}
    Integer                 response status                 400

Create url with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/url                        {"db_id": "toto", "sid": "12", "url":"postgres://user:password@127.0.0.1/db3" }
    Integer                 response status                 400

Modify url should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/url.json
    PUT                     /api/url/16                      {"db_id": 3, "sid": "13", "url":"postgres://user:password@127.0.0.1/db3" }
    Integer                 response status                 200
    String                  response body sid               13
    
Modify url with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/url/16                     {"product_name": "vm-xl-2"}
    Integer                 response status                 400

Modify url with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/url/16                     {"sid": 123}
    Integer                 response status                 400

Modify non existing url should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/url/99                     {"db_id": 3, "sid": "13", "url":"postgres://user:password@127.0.0.1/db3" }
    Integer                 response status                 404

Modify url with invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/url/sdlkfgjh                {"sid": "12"}
    Integer                 response status                 400

Delete existing url should succeed
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/url/16
    Integer                 response status                 200

Delete non-existing url should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/url/44
    Integer                 response status                 404

Delete invalid url_id should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/url/sdkjfgh
    Integer                 response status                 400
