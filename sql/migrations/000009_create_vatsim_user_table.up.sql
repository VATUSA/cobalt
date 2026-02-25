CREATE TABLE vatsim_user (
    cid bigint not null primary key,
    name_first varchar(240) not null,
    name_last varchar(240) not null,
    email varchar(240) not null,
    rating integer not null,
    pilotrating integer not null,
    militaryrating integer not null,
    suspend_date datetime null,
    registration_date datetime null,
    region_id varchar(12) not null,
    division_id varchar(12) not null,
    subdivision_id varchar(12) null,
    latest_rating_change datetime null,
    last_sync datetime not null
);