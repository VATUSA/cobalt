CREATE TABLE facility_title (
    id bigint not null auto_increment primary key,
    facility varchar(4) not null,
    title varchar(128) not null,
    created_at bigint not null,
    unique key uq_facility_title (facility, title)
);

CREATE TABLE user_title (
    id bigint not null auto_increment primary key,
    cid int not null,
    title_id bigint not null,
    grantor_cid int not null,
    granted_at bigint not null,
    unique key uq_user_title (cid, title_id),
    key idx_user_title_title_id (title_id)
);
