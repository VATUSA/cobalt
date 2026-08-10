CREATE TABLE action_log (
    id bigint not null auto_increment primary key,
    actor_cid int not null,
    target_cid int not null,
    action varchar(50) not null,
    log text not null,
    created_at bigint not null,
    updated_at bigint not null
);
