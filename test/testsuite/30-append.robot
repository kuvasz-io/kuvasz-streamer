*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Test cases ***
Insert in append table row 1
    Statement should propagate
    ...    insert into t3(name) values('r1')
    ...    Select id, name, salary from t3 order by id
    ...    Select id, name, salary from t3 where sid='{}' order by id

Insert in append table row 2
    Statement should propagate         
    ...    insert into t3(name) values('r2')
    ...    Select id, name, salary from t3 order by id
    ...    Select id, name, salary from t3 where sid='{}' order by id

Update append table non key attribute
    Statement should propagate
    ...    update t3 set name='x1' where id=1
    ...    select id, name, salary from t3 order by id
    ...    select id, name, salary from t3 where sid='{}' order by id

Update append table key attribute
    Statement should propagate        
    ...    update t3 set id=5 where id=1
    ...    select id, name, salary from t3 order by id
    ...    select id, name, salary from t3 where sid='{}' order by id

Delete from append table
    Statement should not propagate
    ...    delete from t3 where id=5
    ...    select id, name, salary from t3 where sid='{}' order by id

Insert in append table row 1 - no sid
    Single Database Statement should propagate
    ...    insert into d3(name) values('r1')
    ...    Select id, name, salary from d3 order by id
    ...    Select id, name, salary from d3 order by id

Insert in append table row 2 - no sid
    Single Database Statement should propagate         
    ...    insert into d3(name) values('r2')
    ...    Select id, name, salary from d3 order by id
    ...    Select id, name, salary from d3 order by id

Update append table non key attribute - no sid
    Single Database Statement should propagate
    ...    update d3 set name='x1' where id=1
    ...    select id, name, salary from d3 order by id
    ...    select id, name, salary from d3 order by id

Update append table key attribute - no sid
    Single Database Statement should propagate        
    ...    update d3 set id=5 where id=1
    ...    select id, name, salary from d3 order by id
    ...    select id, name, salary from d3 order by id

Delete from append table - no sid
    Single Database Statement should not propagate
    ...    delete from t3 where id=5
    ...    select id, name, salary from t3 order by id
