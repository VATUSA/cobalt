CREATE TABLE transfer_request (
    id bigint not null auto_increment primary key,
    cid bigint not null,
    from_facility varchar(12) not null,
    to_facility varchar(12) not null,
    reason mediumtext not null,
    created_at timestamp not null,
    status varchar(40) not null
);

CREATE TABLE transfer_history (
    id bigint not null auto_increment primary key,
    cid bigint not null,
    from_facility varchar(12) not null,
    to_facility varchar(12) not null,
    reason mediumtext not null,
    requested_at timestamp not null,
    completed_at timestamp not null,
    approver_cid bigint null
);

CREATE TABLE visit_request (
    id bigint not null auto_increment primary key,
    cid bigint not null,
    facility varchar(12) not null,
    reason mediumtext not null,
    created_at timestamp not null,
    is_rejected bool default false
);

CREATE TABLE visit_history (
    id bigint not null auto_increment primary key,
    cid bigint not null,
    facility varchar(12) not null,
    reason mediumtext not null,
    requested_at timestamp not null,
    approved_at timestamp not null,
    approver_cid bigint null,
    removed_at timestamp null
);