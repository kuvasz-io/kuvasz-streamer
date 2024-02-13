create database dest;
\c dest
CREATE TYPE complex AS (
    r       double precision,
    i       double precision
);

create table t1(sid text, id int, name text, salary int, primary key (sid, id));
create table rt2(sid text, id int, name text, salary int, extra text);
create table t3(sid text, id int, name text, salary int, primary key (sid, id));
create table t4(kvsz_id bigserial, sid text, id int, name text, salary int, 
    kvsz_start timestamptz not null default '1900-01-01 00:00:00', 
    kvsz_end timestamptz not null default '9999-01-01 00:00:00', 
    kvsz_deleted boolean not null default false, 
    primary key(sid, id, kvsz_id));
create table t5(
    sid text,
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

create table t6(sid text, t text);

create table t7(sid text, id bigint, ts timestamptz, name text);

create table s1(sid text, id int, name text, salary int, garbage date);
