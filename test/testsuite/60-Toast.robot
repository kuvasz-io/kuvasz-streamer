*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***

Insert in normal row
    Statement should propagate
    ...    insert into t6(id, name) values(1, 'toast')
    ...    Select id, name, longvalue from t6
    ...    Select id, name, longvalue from t6 where sid='{}'

Update with large value
    ${TOASTVALUE}=      Generate Random String  20000
    Statement should propagate
    ...    update t6 set longvalue='${TOASTVALUE}'
    ...    Select id, name, longvalue from t6
    ...    Select id, name, longvalue from t6 where sid='{}'

Update small value
    Statement should propagate
    ...    update t6 set name='no-toast'
    ...    Select id, name, longvalue from t6
    ...    Select id, name, longvalue from t6 where sid='{}'

Delete from TOAST table
    Statement should propagate
    ...    delete from t6
    ...    Select id, name, longvalue from t6
    ...    Select id, name, longvalue from t6 where sid='{}'
