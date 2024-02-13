*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***
Initial sync should work
    Statement should propagate
    ...    Select 1
    ...    Select count(*) from t7
    ...    Select count(*) from t7 where sid='{}'
