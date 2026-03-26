update user set display_name = '' where display_name is null;
update user set controller_rating = 0 where controller_rating is null;
update user set instructor_rating = 0 where instructor_rating is null;
update user set discord_id = '' where discord_id is null;

ALTER TABLE user
    MODIFY COLUMN display_name varchar(240) not null default '';
alter table user
    modify column controller_rating int not null default 0;
alter table user
    modify column instructor_rating int not null default 0;
alter table user
    modify column discord_id varchar(255) not null default '';