ALTER TABLE event ADD COLUMN review_status text;
ALTER TABLE event ADD COLUMN reviewed_by int;
ALTER TABLE event ADD COLUMN reviewed_on bigint;
