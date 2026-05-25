ALTER TABLE event ADD COLUMN review_status varchar(40);
ALTER TABLE event ADD COLUMN reviewed_by int;
ALTER TABLE event ADD COLUMN reviewed_on bigint;
