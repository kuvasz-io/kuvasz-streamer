*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***
Initial sync should work
    Sleep                   ${SLEEP}
    Statement should propagate
    ...    Select 1
    ...    Select count(*) from t0
    ...    Select count(*) from t0 where sid='{}'
