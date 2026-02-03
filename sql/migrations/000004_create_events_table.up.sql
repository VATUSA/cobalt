CREATE TABLE event (
    id bigint not null auto_increment primary key,
    title varchar(240) not null,
    body text not null,
    banner_image_url text not null,
    facility varchar(4) not null,
    start_time bigint not null,
    end_time bigint not null,
    created_at bigint not null,
    created_by int not null,
    updated_at bigint not null,
    updated_by integer not null
)