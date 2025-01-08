*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***

GET existing map entry by id should succeed
    Clear Expectations
    Expect Response Body    ${SCHEMA}/map.json
    Set Headers             ${admin}
    GET                     /api/map/1
    Integer                 response status                 200
    Integer                 response body id                1
    Integer                 response body db_id             1
    String                  response body db_name           db1
    String                  response body name              d1
    String                  response body type              clone

GET non existing map entry by id should fail
    Clear Expectations
    Set Headers             ${admin}
    GET                     /api/map/99
    Integer                 response status                 404

GET map entry by invalid id should fail
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/error.json
    GET                     /api/map/dskjhfkdsjfgh
    Integer                 response status                 400

GET full map should succeed
    Clear Expectations
    Set Headers             ${admin}
    Expect Response Body    ${schema}/maps.json
    GET                     /api/map
    Integer                 response status                 200
    Array                   response body                   minItems=20  maxItems=20

Add database and refresh map
    Prepare db3

    # Create database
    Clear Expectations
    Expect Response Body    ${SCHEMA}/db.json
    Set Headers             ${admin}
    POST                    /api/db                         {"name": "db3"}
    Integer                 response status                 200

    # Create url
    Clear Expectations
    Expect Response Body    ${SCHEMA}/url.json
    Set Headers             ${admin}
    POST                    /api/url                        {"db_id": 3, "sid": "12", "url":"postgres://kuvasz:kuvasz@127.0.0.1:6012/db3" }
    Integer                 response status                 200

Clone non existing table u0

    # Create table
    Switch database         db3
    Execute SQL string      create table u0(id serial, name text)

    Clone table             u0                              20    \
    
Insert row in u0
    Switch Database              db3
    Execute SQL string           insert into u0(name) values('foo')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u0 where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo

Clone existing table u1
    # Create u1 on source and destination
    Switch database              db3
    Execute SQL string           create table u1(id serial, name text)
    Switch database              dest
    Execute SQL string           create table u1(sid text, id int, name text)
    Clone table                  u1                              21    \

Insert row in u1
    Switch Database              db3
    Execute SQL string           insert into u1(name) values('foo')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u1 where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo

Clone non-existing table with pre-existing data u2
    # Create table
    Switch database         db3
    Execute SQL string      create table u2(id serial, name text)
    Execute SQL string      insert into u2(name) values('foo1')
    Execute SQL string      insert into u2(name) values('foo2')
    Execute SQL string      insert into u2(name) values('foo3')
    Execute SQL string      insert into u2(name) values('foo4')
    Clone table             u2                             22      \

Insert row in u2
    Switch Database              db3
    Execute SQL string           insert into u2(name) values('bar')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u2 where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo1
    Should be equal as strings   ${result}[1][0]                  foo2
    Should be equal as strings   ${result}[2][0]                  foo3
    Should be equal as strings   ${result}[3][0]                  foo4
    Should be equal as strings   ${result}[4][0]                  bar

Clone partitioned table u3
    # Create table
    Switch database         db3
    Execute SQL string      create table u3(id int primary key, name text) partition by range(id)
    Execute SQL string      create table u3_0 partition of u3 for values from (0) to (9)
    Execute SQL string      create table u3_1 partition of u3 for values from (10) to (19)
    Execute SQL string      create table u3_2 partition of u3 for values from (20) to (29)
    Execute SQL string      create table u3_3 partition of u3 for values from (30) to (39)
    Clone table             u3                             23       ?partitions_regex=u3_.*

Insert row in u3
    Switch Database              db3
    Execute SQL string           insert into u3(id, name) values(10, 'foo')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u3 where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo

Clone renamed and partitioned table u4
    # Create table
    Switch database         db3
    Execute SQL string      create table u4(id int primary key, name text) partition by range(id)
    Execute SQL string      create table u4_0 partition of u4 for values from (0) to (9)
    Execute SQL string      create table u4_1 partition of u4 for values from (10) to (19)
    Execute SQL string      create table u4_2 partition of u4 for values from (20) to (29)
    Execute SQL string      create table u4_3 partition of u4 for values from (30) to (39)
    Clone table             u4                             24       ?partitions_regex=u4_.*&target=u4p

Insert row in u4
    Switch Database              db3
    Execute SQL string           insert into u4(id, name) values(10, 'foo')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u4p where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo

Clone append, renamed table u5
    # Create table
    Switch database         db3
    Execute SQL string      create table u5(id int primary key, name text)
    Clone table             u5                             25       ?target=u5p&type=append

Insert row in u5
    Switch Database              db3
    Execute SQL string           insert into u5(id, name) values(10, 'foo')
    Sleep                        ${SLEEP}
    Switch Database              dest
    ${result}=                   Query                            select name from u5p where sid='12'
    Should be equal as strings   ${result}[0][0]                  foo

