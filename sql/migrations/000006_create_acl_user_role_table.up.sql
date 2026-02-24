CREATE TABLE acl_user_role (
    id bigint not null auto_increment primary key,
    cid int not null,
    facility varchar(4) not null,
    role varchar(120) not null,
    grantor_cid int not null,
    granted_at bigint not null
);