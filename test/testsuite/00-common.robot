*** Settings ***
Library           DatabaseLibrary
Library           OperatingSystem
Library           Collections
Library           String
Library           REST               http://127.0.0.1:8000

*** Variables ***
${SLEEP}         1.8
@{PGVERSIONS}=   12    13    14    15    16    17
${SOURCE}        12
${SCHEMA}        ../api/schema
${ADMIN}         {"content-type": "application/json"}

*** Keywords ***
Connect To All Databases
    FOR    ${PG}    IN    @{PGVERSIONS}
        Connect To Database Using Custom Connection String   psycopg2    postgres://kuvasz:kuvasz@127.0.0.1:60${PG}/db1?application_name=robot    alias=${PG}
        Execute SQL string    truncate t1 restart identity
        Execute SQL string    truncate t2 restart identity
        Execute SQL string    truncate t3 restart identity
        Execute SQL string    truncate t4 restart identity
        Execute SQL string    truncate t5 restart identity
        Execute SQL string    truncate t6 restart identity
        Execute SQL string    truncate t7 restart identity
        Execute SQL string    truncate t8 restart identity
        Execute SQL string    truncate private.t8 restart identity
        Set Auto Commit
    END
    Switch database            ${SOURCE}
    Execute SQL string        truncate d1 restart identity
    Execute SQL string        truncate d2 restart identity
    Execute SQL string        truncate d3 restart identity
    Execute SQL string        truncate d4 restart identity
    Execute SQL string        truncate d6 restart identity
    Execute SQL string        truncate d7 restart identity
    Execute SQL string        truncate d8 restart identity
    Execute SQL string        truncate private.d8 restart identity

    Connect To Database Using Custom Connection String       psycopg2    postgres://kuvasz:kuvasz@127.0.0.1:6012/dest?application_name=robot    alias=dest
    Switch database           dest
    Execute SQL string        truncate t1 restart identity
    Execute SQL string        truncate rt2 restart identity
    Execute SQL string        truncate t3 restart identity
    Execute SQL string        truncate t4 restart identity
    Execute SQL string        truncate t5 restart identity
    Execute SQL string        truncate t6 restart identity
    Execute SQL string        truncate t7 restart identity
    Execute SQL string        truncate t8 restart identity
    Execute SQL string        truncate pt8 restart identity
    Execute SQL string        truncate d1 restart identity
    Execute SQL string        truncate rd2 restart identity
    Execute SQL string        truncate d3 restart identity
    Execute SQL string        truncate d4 restart identity
    Execute SQL string        truncate d6 restart identity
    Execute SQL string        truncate d7 restart identity
    Execute SQL string        truncate d8 restart identity
    Execute SQL string        truncate pd8 restart identity
    Set Auto Commit

Prepare db3
    Switch database           12
    Execute SQL string        drop database if exists db3
    Execute SQL string        create database db3
    Switch database           dest
    Execute SQL string        drop table if exists u0
    Execute SQL string        drop table if exists u1
    Execute SQL string        drop table if exists u2
    Execute SQL string        drop table if exists u3    
    Connect To Database       psycopg2    db3    kuvasz    kuvasz    127.0.0.1    6012    alias=db3
    Switch database           db3
    Set Auto Commit

Statement should propagate
    [Arguments]             ${ACTION}  ${TEMPLATEDSOURCEQUERY}    ${TEMPLATEDESTQUERY}
    FOR    ${PG}    IN    @{PGVERSIONS}
        ${SOURCEQUERY}=         Format string    ${TEMPLATEDSOURCEQUERY}    ${PG}
        ${DESTQUERY}=           Format string    ${TEMPLATEDESTQUERY}    ${PG}
        Switch Database         ${PG}
        Execute SQL string      ${ACTION}
        ${src}=                 Query            ${SOURCEQUERY}
        Sleep                   ${SLEEP}
        Switch Database         dest
        ${dest}=                Query            ${DESTQUERY}
        Lists Should Be Equal   ${src}           ${dest}
    END

Statement should not propagate
    [Arguments]             ${ACTION}  ${TEMPLATEQUERY}
    FOR    ${PG}    IN    @{PGVERSIONS}
        ${DESTQUERY}=           Format string    ${TEMPLATEQUERY}    ${PG}
        Switch Database         dest
        ${before}=              Query            ${DESTQUERY}
        Switch Database         ${PG}
        Execute SQL string      ${ACTION}
        Sleep                   ${SLEEP}
        Switch Database         dest
        ${after}=               Query             ${DESTQUERY}
        Lists Should Be Equal   ${before}         ${after}
    END

Single database statement should propagate
    [Arguments]             ${ACTION}  ${SOURCEQUERY}    ${DESTQUERY}
    Switch Database         ${SOURCE}
    Execute SQL string      ${ACTION}
    ${src}=                 Query            ${SOURCEQUERY}
    Sleep                   ${SLEEP}
    Switch Database         dest
    ${dest}=                Query            ${DESTQUERY}
    Lists Should Be Equal   ${src}           ${dest}

Single database Statement should not propagate
    [Arguments]             ${ACTION}  ${QUERY}
    Switch Database         dest
    ${before}=              Query            ${QUERY}
    Switch Database         ${SOURCE}
    Execute SQL string      ${ACTION}
    Sleep                   ${SLEEP}
    Switch Database         dest
    ${after}=               Query             ${QUERY}
    Lists Should Be Equal   ${before}         ${after}

Execute on source
    [Arguments]             ${SQL}
    FOR    ${PG}    IN    @{PGVERSIONS}
        Switch Database         ${PG}
        Execute SQL string      ${SQL}
    END

Clone table
    [Arguments]                  ${table}    ${id}      ${param}
    # Refresh map
    Clear Expectations     
    Set Headers                  ${admin}
    POST                         /api/map/refresh
    Integer                      response status                 200

    # Check map was updated
    Clear Expectations     
    Expect Response Body         ${SCHEMA}/map.json
    Set Headers                  ${admin}
    GET                          /api/map/${id}
    Integer                      response status                 200
    String                       response body name              ${table}

    # Clone table
    Clear Expectations     
    Expect Response Body         ${SCHEMA}/map.json
    Set Headers                  ${admin}
    POST                         /api/map/${id}/clone${param}                
    Integer                      response status                 200

    # Check map
    Clear Expectations     
    Expect Response Body         ${SCHEMA}/map.json
    Set Headers                  ${admin}
    GET                          /api/map/${id}
    Integer                      response status                 200
    Integer                      response body id                ${id}
    Integer                      response body db_id             3
    String                       response body db_name           db3
    String                       response body name              ${table}
    String                       response body type              clone    append    history
    Boolean                      response body replicated        true
    Boolean                      response body present           true

    # Restart engine to apply config
    Clear Expectations     
    Set Headers                  ${admin}
    POST                         /api/url/restart
    Sleep                        5
