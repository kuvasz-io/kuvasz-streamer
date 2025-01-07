*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***

Insert in partition 0
    Statement should propagate
    ...    insert into t7(id, name) values(0, 'p0')
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

Insert in partition 1
    Statement should propagate
    ...    insert into t7(id, name) values(10, 'p1')
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

Insert in partition 2
    Statement should propagate
    ...    insert into t7(id, name) values(20, 'p2')
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

Insert in partition 3
    Statement should propagate
    ...    insert into t7(id, name) values(30, 'p3')
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

Update partition 2
    Statement should propagate
    ...    update t7 set name='p2x' where id=20
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

Delete from partition 3
    Statement should propagate
    ...    delete from t7 where id=30
    ...    Select id, name from t7 order by id
    ...    Select id, name from t7 where sid='{}' order by id

# No SID

Insert in partition 0 - no sid
    Single Database Statement should propagate
    ...    insert into d7(id, name) values(0, 'p0')
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id

Insert in partition 1 - no sid
    Single Database Statement should propagate
    ...    insert into d7(id, name) values(10, 'p1')
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id

Insert in partition 2 - no sid
    Single Database Statement should propagate
    ...    insert into d7(id, name) values(20, 'p2')
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id

Insert in partition 3 - no sid
    Single Database Statement should propagate
    ...    insert into d7(id, name) values(30, 'p3')
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id

Update partition 2 - no sid
    Single Database Statement should propagate
    ...    update d7 set name='p2x' where id=20
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id

Delete from partition 3 - no sid
    Single Database Statement should propagate
    ...    delete from d7 where id=30
    ...    Select id, name from d7 order by id
    ...    Select id, name from d7 order by id