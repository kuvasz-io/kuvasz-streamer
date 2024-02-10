*** Settings ***
Library           DatabaseLibrary
Library           OperatingSystem
Library           Collections
Library           String

*** Variables ***
${SLEEP}          0.8
@{PGVERSIONS}=    12    13    14    15    16

*** Keywords ***
Connect To All Databases
    FOR    ${PG}    IN    @{PGVERSIONS}
        Connect To Database    psycopg2    db1    kuvasz    kuvasz    127.0.0.1    60${PG}    alias=${PG}
        Set Auto Commit
        Execute SQL string    truncate t1 restart identity
        Execute SQL string    truncate t2 restart identity
        Execute SQL string    truncate t3 restart identity
        Execute SQL string    truncate t4 restart identity
        Execute SQL string    truncate t5 restart identity
        Execute SQL string    truncate t6 restart identity
    END
    Connect To Database    psycopg2    dest   kuvasz    kuvasz    127.0.0.1    6012    alias=dest
    Execute SQL string    truncate t1 restart identity
    Execute SQL string    truncate rt2 restart identity
    Execute SQL string    truncate t3 restart identity
    Execute SQL string    truncate t4 restart identity
    Execute SQL string    truncate t5 restart identity
    Execute SQL string    truncate t6 restart identity
    Set Auto Commit
    
Statement should propagate
    [Arguments]             ${ACTION}  ${SOURCEQUERY}    ${TEMPLATEQUERY}
    FOR    ${PG}    IN    @{PGVERSIONS}
        ${DESTQUERY}=           Format string    ${TEMPLATEQUERY}    ${PG}
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
