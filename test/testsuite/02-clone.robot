*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***
Insert in clone table row 1
    Statement should propagate
    ...    insert into t1(name) values('r1')
    ...    Select id, name, salary from t1 order by id
    ...    Select id, name, salary from t1 where sid='{}' order by id

Insert in clone table row 2
    Statement should propagate         
    ...    insert into t1(name) values('r2')
    ...    Select id, name, salary from t1 order by id
    ...    Select id, name, salary from t1 where sid='{}' order by id

Insert in clone table with RIF
    Statement should propagate         
    ...    insert into t2(id, name) values(1, 'r1')
    ...    Select id, name, salary from t2 order by id, name, salary
    ...    Select id, name, salary from rt2 where sid='{}' order by id, name, salary

Insert in clone table with RIF non-null attribute
    Statement should propagate         
    ...    insert into t2(id, name, extra) values(1, 'r1', 'foo')
    ...    Select id, name, salary from t2 order by id, name, salary
    ...    Select id, name, salary from rt2 where sid='{}' order by id, name, salary

Update clone table non key attribute
    Statement should propagate
    ...    update t1 set name='x1' where id=1
    ...    select id, name, salary from t1 order by id
    ...    select id, name, salary from t1 where sid='{}' order by id

Update clone table key attribute
    Statement should propagate        
    ...    update t1 set id=5 where id=1
    ...    select id, name, salary from t1 order by id
    ...    select id, name, salary from t1 where sid='{}'

Update clone table with RIF - multiple rows
    Statement should propagate
    ...    update t2 set id=5 where id=1
    ...    select id, name, salary from t2 order by id, name, salary
    ...    select id, name, salary from rt2 where sid='{}' order by id, name, salary

Update clone table with RIF - salary
    Statement should propagate
    ...    update t2 set salary=2 where id=5 and extra = 'foo'
    ...    select id, name, salary from t2 order by id, name, salary
    ...    select id, name, salary from rt2 where sid='{}' order by id, name, salary

Update clone table with RIF - extra
    Statement should propagate
    ...    update t2 set salary=3, extra='bar' where extra is null
    ...    select id, name, salary from t2 order by id, name, salary
    ...    select id, name, salary from rt2 where sid='{}' order by id, name, salary

Delete from clone table
    Statement should propagate
    ...    delete from t1 where id=5
    ...    select id, name, salary from t1 order by id
    ...    select id, name, salary from t1 where sid='{}' order by id

Delete from clone table with RIF
    Statement should propagate
    ...    delete from t2 where id=5
    ...    select id, name, salary from t2 order by id, name, salary
    ...    select id, name, salary from rt2 where sid='{}' order by id, name, salary
