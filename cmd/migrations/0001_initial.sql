create table db(
    db_id integer primary key, 
    name text not null unique);

create table url(
    url_id  integer primary key,
    db_id   integer not null references db(db_id), 
    url     text not null, 
    sid     text not null,
    unique (db_id, sid)
);

create table tbl(
    tbl_id           integer primary key,
    db_id            integer not null references db(db_id), 
    name             text    not null, 
    type             text    not null,
    target           text    not null,
    partitions_regex text    null,
    unique (db_id, name)
);

pragma foreign_keys=ON;

