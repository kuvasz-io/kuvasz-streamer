*** Settings ***
Resource      00-common.robot

*** Test cases ***

GET existing db by id should succeed
    Clear Expectations
    Expect Response Body    ${SCHEMA}/db.json
    Set Headers             ${admin}
    GET                     /api/db/1
    Integer                 response status                 200
    Integer                 response body id                1
    String                  response body name              db1

GET non existing db by id should fail
    Clear Expectations
    Set Headers             ${admin}
    GET                     /api/db/99
    Integer                 response status                 404

GET db by invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    GET                     /api/db/dskjhfkdsjfgh
    Integer                 response status                 400

GET all dbs should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/dbs.json
    GET                     /api/db
    Integer                 response status                 200
    Array                   response body                   minItems=2  maxItems=2

Create db should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/db.json
    POST                    /api/db                         {"name": "db3"}
    Integer                 response status                 200
    String                  response body name              db3    
    
Create db with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/db                         {"product_name": "vm-xl-2"}
    Integer                 response status                 400

Create db with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/db                         {"name": 123}
    Integer                 response status                 400

Modify db should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/db.json
    PUT                     /api/db/3                       {"name": "newdb3"}
    Integer                 response status                 200
    String                  response body name              newdb3    
    
Modify with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/db/3                       {"product_name": "vm-xl-2"}
    Integer                 response status                 400

Modify db with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/db/3                        {"name": 123}
    Integer                 response status                 400

Modify non existing db should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/db/4                        {"name": "newdb4"}
    Integer                 response status                 404

Modify db with invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/db/sdlkfgjh                {"name": "newdb4"}
    Integer                 response status                 400

Delete existing db should succeed
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/db/3
    Integer                 response status                 200

Delete non-existing db should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/db/44
    Integer                 response status                 404

Delete invalid db_id should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/db/sdkjfgh
    Integer                 response status                 400
