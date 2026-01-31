CREATE TABLE actor (
    id integer not null auto_increment primary key,
    name varchar(120) not null unique,
    actor_type varchar(40) not null,
    facility_owner varchar(4) null,
    cid_owner integer null,
    rate_limit_override integer not null default 0,
    rate_limit_bypass boolean not null default false,
    is_active boolean not null default false,
    created_at bigint not null,
    created_by_cid integer not null
);

CREATE TABLE actor_acl (
    id bigint not null auto_increment primary key,
    actor_id integer not null,
    acl varchar(120) not null,
    scope_facility varchar(4) not null,
    created_at bigint not null,
    created_by_cid integer not null,
    FOREIGN KEY (actor_id) REFERENCES actor (id) ON DELETE CASCADE
);

CREATE TABLE actor_token (
    id integer not null auto_increment primary key,
    actor_id integer not null,
    comment varchar(240) null,
    token varchar(120) not null,
    is_active boolean not null default false,
    created_at bigint not null,
    created_by_cid integer not null,
    updated_at bigint not null,
    updated_by_cid integer not null,
    FOREIGN KEY (actor_id) REFERENCES actor (id) ON DELETE CASCADE
);