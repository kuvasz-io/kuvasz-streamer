insert into t1(name, salary, garbage) values('toto', 123, '2023-01-01');
insert into t2(id, name, salary, extra, garbage) values(floor(random()*1000000)::int, 'toto', 123, 'extra', '2023-01-01');
-- insert into t3(id serial primary key, name text, salary int, garbage date);
-- create table t4(id serial primary key, ts_created timestamptz default now(), ts_updated timestamptz default now(), name text, salary int, garbage date);
insert into t5 values(
-9023372036854770000, -- bigint
1, -- bigserial
'1', -- bit
'110101', -- bit varying
true, -- boolean
'(1,1),(4,4)', -- box
'\xdeadbeef', -- bytea
'A', -- character
'ABCD', -- character varying
'192.168.0.1', -- cidr
'(2,2),4', -- circle
'2023-01-01', -- date
123.123456789012345, -- double precision
'192.168.0.0/16', -- inet
1000000, -- integer
'P1DT5M', -- interval
'{"name":"value"}', -- json
'{"name":"value"}', -- jsonb
'{1,2,3}', -- line
'(1,1),(5,5)', -- lseg
'08:00:2b:01:02:03', -- macaddr
'08:00:2b:01:02:03:04:05', -- macaddr8
123.12, -- money
1234567890.12345678901234567890, -- numeric
'[(1,1),(2,1),(4,4)]', -- path
'16/B374D848', -- pg_lsn
'(1,2)', -- point
'(1,1),(2,1),(4,4)', -- polygon
123.123456, -- real
32000, -- smallint
1, -- smallserial
1, -- serial
'abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij', -- text
'12:34:56.123', -- time
'12:34:56.123+02', -- time with time zone
'2023-01-02 01:02:03.123', -- timestamp
'2023-01-02 01:02:03.123+02', -- timestamp with time zone
'fat & rat', -- tsquery
'a fat cat sat on a mat and ate a fat rat', -- tsvector
'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- uuid
'<foo>bar</foo>', -- xml    
'{1,2,3,4}',
'(1,2)'
);

insert into t6 values(
    sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)
  ||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea)||sha512(random()::text::bytea));