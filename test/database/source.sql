create database db1;
\c db1
CREATE TYPE complex AS (
    r       double precision,
    i       double precision
);

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
create table t6(t text);
create table t7(id bigserial, ts timestamptz default now(), name text);

create publication kuvasz_db1 for table t1,t2,t3,t4,t5,t6,t7;

create database db2;
\c db2
create table s1(id serial primary key, name text, salary int, garbage date);
create publication kuvasz_db2 for table s1;
