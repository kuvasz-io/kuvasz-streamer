*** Settings ***
Resource      00-common.robot

*** Test cases ***

GET existing tbl by id should succeed
    Clear Expectations
    Expect Response Body    ${SCHEMA}/tbl.json
    Set Headers             ${admin}
    GET                     /api/tbl/1
    Integer                 response status                 200
    Integer                 response body id                1
    Integer                 response body db_id             1
    String                  response body db_name           db1

GET non existing tbl by id should fail
    Clear Expectations
    Set Headers             ${admin}
    GET                     /api/tbl/99
    Integer                 response status                 404

GET tbl by invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    GET                     /api/tbl/dskjhfkdsjfgh
    Integer                 response status                 400

GET all tbls should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/tbls.json
    GET                     /api/tbl
    Integer                 response status                 200
    Array                   response body                   minItems=20  maxItems=20

Create tbl should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/tbl.json
    POST                    /api/tbl                        {"db_id":3,"name":"foo","type":"clone","target":"blah"} 
    Integer                 response status                 200
    Integer                 response body db_id             3

Create tbl with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/tbl                        {"db_id": 3, "product_name": "vm-xl-2"}
    Integer                 response status                 400

Create tbl with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    POST                    /api/tbl                        {"db_id": "toto", "sid": "12", "tbl":"postgres://user:password@127.0.0.1/db3" }
    Integer                 response status                 400

Modify tbl should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/tbl.json
    PUT                     /api/tbl/12                     {"db_id":3,"schema":"public","name":"bar","type":"clone","target":"bar"}
    Integer                 response status                 200
    String                  response body name              bar
    
Modify tbl with missing parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/tbl/12                     {"db_id":3,"name":"bar","type":"clone"}
    Integer                 response status                 400

Modify tbl with invalid parameters should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/tbl/12                     {"db_id":3,"schema":"public","name":123,"type":"clone","target":"bar"}
    Integer                 response status                 400

Modify non existing tbl should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/tbl/99                     {"db_id":3,"schema":"public","name":"bar","type":"clone","target":"bar"}
    Integer                 response status                 404

Modify tbl with invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    PUT                     /api/tbl/sdlkfgjh               {"db_id":3,"name":"bar","type":"clone","target":"bar"}
    Integer                 response status                 400

Delete existing tbl should succeed
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/tbl/12
    Integer                 response status                 200

Delete non-existing tbl should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/tbl/44
    Integer                 response status                 404

Delete invalid tbl_id should fail
    Clear Expectations
    Set Headers             ${admin}
    DELETE                  /api/tbl/sdkjfgh
    Integer                 response status                 400
