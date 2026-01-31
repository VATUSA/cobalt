CREATE TABLE role_global (
    id bigint not null auto_increment primary key,
    cid integer not null,
    role varchar(40) not null,
    created_at bigint not null,
    created_by integer not null
);

CREATE TABLE role_facility (
    id bigint not null auto_increment primary key,
    cid integer not null,
    role varchar(40) not null,
    facility varchar(40) not null,
    created_at bigint not null,
    created_by integer not null
);