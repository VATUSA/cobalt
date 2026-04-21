CREATE TABLE user_rating_hours (
    cid bigint not null,
    rating int not null,
    hours int not null,
    last_check_time timestamp not null,
    PRIMARY KEY (cid, rating)
);