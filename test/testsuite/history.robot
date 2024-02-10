*** Settings ***
Resource           common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Variables ***
${DESTQUERY}       Select id, name, salary, kvsz_start, kvsz_end, kvsz_deleted from t4 where sid='{}' and id=1 order by id, kvsz_start

*** Test cases ***
Insert in history table row 1
    Statement should propagate
    ...    insert into t4(name) values('r1')
    ...    Select id, name, salary, '1900-01-01'::timestamptz, '9999-01-01'::timestamptz, false from t4 order by id
    ...    Select id, name, salary, kvsz_start, kvsz_end, kvsz_deleted from t4 where sid='{}' order by id

Insert in history table row 2
    Statement should propagate         
    ...    insert into t4(name) values('r2')
    ...    Select id, name, salary, '1900-01-01'::timestamptz, '9999-01-01'::timestamptz, false from t4 order by id
    ...    Select id, name, salary, kvsz_start, kvsz_end, kvsz_deleted from t4 where sid='{}' order by id

Update history table row 1 - take 1
    FOR    ${PG}    IN    @{PGVERSIONS}
        Switch Database         ${PG}
        Execute SQL string      update t4 set name='x1' where id=1
        Sleep                   ${SLEEP}
        Switch Database         dest
        ${query}=               Format string    ${DESTQUERY}    ${PG}
        ${dest}=                Query            ${query}
        Should Be Equal As Strings  ${dest}[0][0]  1
        Should Be Equal As Strings  ${dest}[0][1]  r1
        Should Be Equal As Strings  ${dest}[0][2]  None
        Should Be Equal As Strings  ${dest}[0][3]  1900-01-01 00:00:00+00:00
        Should Be Equal As Strings  ${dest}[0][5]  False
        Should Be Equal As Strings  ${dest}[1][0]  1
        Should Be Equal As Strings  ${dest}[1][1]  x1
        Should Be Equal As Strings  ${dest}[1][2]  None
        Should Be Equal As Strings  ${dest}[1][3]  ${dest}[0][4]
        Should Be Equal As Strings  ${dest}[1][4]  9999-01-01 00:00:00+00:00
        Should Be Equal As Strings  ${dest}[1][5]  False
    END

Update history table row 1 - take 2
    FOR    ${PG}    IN    @{PGVERSIONS}
        Switch Database         ${PG}
        Execute SQL string      update t4 set name='z1' where id=1
        Sleep                   ${SLEEP}
        Switch Database         dest
        ${query}=               Format string    ${DESTQUERY}    ${PG}
        ${dest}=                Query            ${query}
        Should Be Equal As Strings  ${dest}[0][0]  1
        Should Be Equal As Strings  ${dest}[0][1]  r1
        Should Be Equal As Strings  ${dest}[0][2]  None
        Should Be Equal As Strings  ${dest}[0][3]  1900-01-01 00:00:00+00:00
        Should Be Equal As Strings  ${dest}[0][4]  ${dest}[1][3]
        Should Be Equal As Strings  ${dest}[0][5]  False
        Should Be Equal As Strings  ${dest}[1][0]  1
        Should Be Equal As Strings  ${dest}[1][1]  x1
        Should Be Equal As Strings  ${dest}[1][2]  None
        Should Be Equal As Strings  ${dest}[1][3]  ${dest}[0][4]
        Should Be Equal As Strings  ${dest}[1][4]  ${dest}[2][3]
        Should Be Equal As Strings  ${dest}[1][5]  False
        Should Be Equal As Strings  ${dest}[2][0]  1
        Should Be Equal As Strings  ${dest}[2][1]  z1
        Should Be Equal As Strings  ${dest}[2][2]  None
        Should Be Equal As Strings  ${dest}[2][3]  ${dest}[1][4]
        Should Be Equal As Strings  ${dest}[2][4]  9999-01-01 00:00:00+00:00
        Should Be Equal As Strings  ${dest}[2][5]  False
    END

Delete history table row 1
    FOR    ${PG}    IN    @{PGVERSIONS}
        Switch Database                 ${PG}
        Execute SQL string              delete from t4 where id=1
        Sleep                           ${SLEEP}
        Switch Database                 dest
        ${query}=                       Format string  ${DESTQUERY}    ${PG}
        ${dest}=                        Query          ${query}
        Should Be Equal As Strings      ${dest}[0][0]  1
        Should Be Equal As Strings      ${dest}[0][1]  r1
        Should Be Equal As Strings      ${dest}[0][2]  None
        Should Be Equal As Strings      ${dest}[0][3]  1900-01-01 00:00:00+00:00
        Should Be Equal As Strings      ${dest}[0][4]  ${dest}[1][3]
        Should Be Equal As Strings      ${dest}[0][5]  False
        Should Be Equal As Strings      ${dest}[1][0]  1
        Should Be Equal As Strings      ${dest}[1][1]  x1
        Should Be Equal As Strings      ${dest}[1][2]  None
        Should Be Equal As Strings      ${dest}[1][3]  ${dest}[0][4]
        Should Be Equal As Strings      ${dest}[1][4]  ${dest}[2][3]
        Should Be Equal As Strings      ${dest}[1][5]  False
        Should Be Equal As Strings      ${dest}[2][0]  1
        Should Be Equal As Strings      ${dest}[2][1]  z1
        Should Be Equal As Strings      ${dest}[2][2]  None
        Should Be Equal As Strings      ${dest}[2][3]  ${dest}[1][4]
        Should Not Be Equal As Strings  ${dest}[2][4]  9999-01-01 00:00:00+00:00
        Should Be Equal As Strings      ${dest}[2][5]  True
    END
