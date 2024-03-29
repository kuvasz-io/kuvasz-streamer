insert into db(db_id, name) values(1, 'db1');

insert into url(url_id, db_id, url, sid) values(1, 1, 'postgres://kuvasz:kuvasz@127.0.0.1:6012/db1?replication=database&application_name=repl_db1', '12');
insert into url(url_id, db_id, url, sid) values(2, 1, 'postgres://kuvasz:kuvasz@127.0.0.1:6013/db1?replication=database&application_name=repl_db1', '13');
insert into url(url_id, db_id, url, sid) values(3, 1, 'postgres://kuvasz:kuvasz@127.0.0.1:6014/db1?replication=database&application_name=repl_db1', '14');
insert into url(url_id, db_id, url, sid) values(4, 1, 'postgres://kuvasz:kuvasz@127.0.0.1:6015/db1?replication=database&application_name=repl_db1', '15');
insert into url(url_id, db_id, url, sid) values(5, 1, 'postgres://kuvasz:kuvasz@127.0.0.1:6016/db1?replication=database&application_name=repl_db1', '16');

insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(1, 1,'t0', 'clone',  't0',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(2, 1,'t1', 'clone',  't1',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(3, 1,'t2', 'clone',  'rt2', NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(4, 1,'t3', 'append', 't3',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(5, 1,'t4', 'history','t4',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(6, 1,'t5', 'clone',  't5',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(7, 1,'t6', 'clone',  't6',  NULL);
insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(8, 1,'t7', 'clone',  't7',  't7_.*');

insert into db(db_id, name) values(2, 'db2');

insert into url(url_id, db_id, url, sid) values(6, 2, 'postgres://kuvasz:kuvasz@127.0.0.1:6012/db2?replication=database&application_name=repl_db2', '12');
insert into url(url_id, db_id, url, sid) values(7, 2, 'postgres://kuvasz:kuvasz@127.0.0.1:6013/db2?replication=database&application_name=repl_db2', '13');
insert into url(url_id, db_id, url, sid) values(8, 2, 'postgres://kuvasz:kuvasz@127.0.0.1:6014/db2?replication=database&application_name=repl_db2', '14');
insert into url(url_id, db_id, url, sid) values(9, 2, 'postgres://kuvasz:kuvasz@127.0.0.1:6015/db2?replication=database&application_name=repl_db2', '15');
insert into url(url_id, db_id, url, sid) values(10,2, 'postgres://kuvasz:kuvasz@127.0.0.1:6016/db2?replication=database&application_name=repl_db2', '16');

insert into tbl(tbl_id, db_id, name, type, target, partitions_regex) values(9,  2,'s1', 'clone',  's1',  NULL);
