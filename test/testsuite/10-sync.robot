*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***
Initial sync should work
    Sleep                   ${SLEEP}
    Statement should propagate
    ...    Select 1
    ...    Select '{}', * from t0 order by id
    ...    Select * from t0 where sid='{}' order by id

Initial sync should work - no sid
    Single database statement should propagate
    ...    Select 1
    ...    Select * from d0 order by id
    ...    Select * from d0 order by id
