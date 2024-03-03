CREATE TABLE public.pgbench_accounts (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
PARTITION BY RANGE (aid);

CREATE TABLE public.pgbench_accounts_1 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');


CREATE TABLE public.pgbench_accounts_10 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_2 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_3 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_4 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_5 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_6 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_7 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_8 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_accounts_9 (
    sid text,
    aid integer NOT NULL,
    bid integer,
    abalance integer,
    filler character(84)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_branches (
    sid text,
    bid integer NOT NULL,
    bbalance integer,
    filler character(88)
)
WITH (fillfactor='100');

CREATE TABLE public.pgbench_history (
    sid text,
    tid integer,
    bid integer,
    aid integer,
    delta integer,
    mtime timestamp without time zone,
    filler character(22)
);

CREATE TABLE public.pgbench_tellers (
    sid text,
    tid integer NOT NULL,
    bid integer,
    tbalance integer,
    filler character(84)
)
WITH (fillfactor='100');

ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_1 FOR VALUES FROM (MINVALUE) TO (10000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_10 FOR VALUES FROM (90000001) TO (MAXVALUE);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_2 FOR VALUES FROM (10000001) TO (20000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_3 FOR VALUES FROM (20000001) TO (30000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_4 FOR VALUES FROM (30000001) TO (40000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_5 FOR VALUES FROM (40000001) TO (50000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_6 FOR VALUES FROM (50000001) TO (60000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_7 FOR VALUES FROM (60000001) TO (70000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_8 FOR VALUES FROM (70000001) TO (80000001);
ALTER TABLE ONLY public.pgbench_accounts ATTACH PARTITION public.pgbench_accounts_9 FOR VALUES FROM (80000001) TO (90000001);
ALTER TABLE ONLY public.pgbench_accounts
    ADD CONSTRAINT pgbench_accounts_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_10
    ADD CONSTRAINT pgbench_accounts_10_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_1
    ADD CONSTRAINT pgbench_accounts_1_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_2
    ADD CONSTRAINT pgbench_accounts_2_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_3
    ADD CONSTRAINT pgbench_accounts_3_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_4
    ADD CONSTRAINT pgbench_accounts_4_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_5
    ADD CONSTRAINT pgbench_accounts_5_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_6
    ADD CONSTRAINT pgbench_accounts_6_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_7
    ADD CONSTRAINT pgbench_accounts_7_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_8
    ADD CONSTRAINT pgbench_accounts_8_pkey PRIMARY KEY (aid);
ALTER TABLE ONLY public.pgbench_accounts_9
    ADD CONSTRAINT pgbench_accounts_9_pkey PRIMARY KEY (aid);

ALTER TABLE ONLY public.pgbench_branches
    ADD CONSTRAINT pgbench_branches_pkey PRIMARY KEY (bid);

ALTER TABLE ONLY public.pgbench_tellers
    ADD CONSTRAINT pgbench_tellers_pkey PRIMARY KEY (tid);

ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_10_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_1_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_2_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_3_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_4_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_5_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_6_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_7_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_8_pkey;
ALTER INDEX public.pgbench_accounts_pkey ATTACH PARTITION public.pgbench_accounts_9_pkey;
