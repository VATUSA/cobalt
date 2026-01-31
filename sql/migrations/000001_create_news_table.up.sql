CREATE TABLE news_post (
    id integer not null auto_increment primary key,
    title varchar(120) not null,
    body text not null,
    author_cid integer not null,
    post_time bigint not null
);