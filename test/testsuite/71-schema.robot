*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases

*** Variables ***
${DESTQUERY}       Select id, name, salary, kvsz_start, kvsz_end, kvsz_deleted from t4 where sid='{}' and id=1 order by id, kvsz_start

*** Test cases ***
Insert in public table
    Statement should propagate
    ...    insert into t8(name) values('p1')
    ...    Select id, name from t8 order by id
    ...    Select id, name from t8 where sid='{}' order by id

Insert in private table
    Statement should propagate
    ...    insert into private.t8(name) values('x1')
    ...    Select id, name from private.t8 order by id
    ...    Select id, name from pt8 where sid='{}' order by id

Update public table non key attribute
    Statement should propagate
    ...    update t8 set name='x1' where id=1
    ...    select id, name from t8 order by id
    ...    select id, name from t8 where sid='{}' order by id

Update private table non key attribute
    Statement should propagate
    ...    update private.t8 set name='z1' where id=1
    ...    select id, name from private.t8 order by id
    ...    select id, name from pt8 where sid='{}' order by id

Update public table key attribute
    Statement should propagate        
    ...    update t8 set id=5 where id=1
    ...    select id, name from t8 order by id
    ...    select id, name from t8 where sid='{}'

Update private table key attribute
    Statement should propagate        
    ...    update private.t8 set id=10 where id=1
    ...    select id, name from private.t8 order by id
    ...    select id, name from pt8 where sid='{}'

# no sid

Insert in public table - no sid
    Single Database Statement should propagate
    ...    insert into d8(name) values('p1')
    ...    Select id, name from d8 order by id
    ...    Select id, name from d8 order by id

Insert in private table - no sid
    Single Database Statement should propagate
    ...    insert into private.d8(name) values('x1')
    ...    Select id, name from private.d8 order by id
    ...    Select id, name from pd8 order by id

Update public table non key attribute - no sid
    Single Database Statement should propagate
    ...    update d8 set name='x1' where id=1
    ...    select id, name from d8 order by id
    ...    select id, name from d8 order by id

Update private table non key attribute - no sid
    Single Database Statement should propagate
    ...    update private.d8 set name='z1' where id=1
    ...    select id, name from private.d8 order by id
    ...    select id, name from pd8 order by id

Update public table key attribute - no sid
    Single Database Statement should propagate        
    ...    update d8 set id=5 where id=1
    ...    select id, name from d8 order by id
    ...    select id, name from d8 order by id

Update private table key attribute - no sid
    Single Database Statement should propagate        
    ...    update private.d8 set id=10 where id=1
    ...    select id, name from private.d8 order by id
    ...    select id, name from pd8 order by id
