CREATE TABLE user (
    cid bigint not null primary key,
    display_name varchar(240) null,
    controller_rating int null,
    instructor_rating int null,
    facility varchar(4) not null,
    discord_id varchar(255) null,
    last_promotion_time timestamp null,
    last_transfer_time timestamp null,
    last_competency_date timestamp null
);

CREATE TABLE user_visit (
    cid bigint not null primary key,
    facility varchar(4) not null
);