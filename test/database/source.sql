create database db1;
\c db1
CREATE TYPE complex AS (
    r       double precision,
    i       double precision
);
create table t0(id bigserial, ts timestamptz default now(), name text);
create table t1(id serial primary key, name text, salary int, garbage date);
create table t2(id int, name text, salary int, extra text);
alter table t2 replica identity full;
create table t3(id serial primary key, name text, salary int, garbage date);
create table t4(id serial primary key, name text, salary int, garbage date);
create table t5(
    f1 bigint,
    f2 bigserial,
    f3 bit,
    f4 bit varying,
    f5 boolean,
    f6 box,
    f7 bytea,
    f8 character,
    f9 character varying,
    f10 cidr,
    f11 circle,
    f12 date,
    f13 double precision,
    f14 inet,
    f15 integer,
    f16 interval,
    f17 json,
    f18 jsonb,
    f19 line,
    f20 lseg,
    f21 macaddr,
    f22 macaddr8,
    f23 money,
    f24 numeric,
    f25 path,
    f26 pg_lsn,
    f28 point,
    f29 polygon,
    f30 real,
    f31 smallint,
    f32 smallserial,
    f33 serial,
    f34 text,
    f35 time,
    f36 time with time zone,
    f37 timestamp,
    f38 timestamp with time zone,
    f39 tsquery,
    f40 tsvector,
    f42 uuid,
    f43 xml,
    f44 integer[],
    f45 complex
    );
alter table t5 replica identity full;
create table t6(id int primary key, name text, longvalue text);
create table t7(id int primary key, name text) partition by range(id);
create table t7_0 partition of t7 for values from (0) to (9);
create table t7_1 partition of t7 for values from (10) to (19);
create table t7_2 partition of t7 for values from (20) to (29);
create table t7_3 partition of t7 for values from (30) to (39);

create database db2;
\c db2
create table s1(id serial primary key, name text, salary int, garbage date);
