CREATE TABLE solo_cert (
    id bigint not null auto_increment primary key,
    cid bigint not null,
    facility varchar(12) not null,
    position varchar(20) not null,
    expires date not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

CREATE INDEX idx_solo_cert_cid ON solo_cert (cid);
CREATE INDEX idx_solo_cert_facility ON solo_cert (facility);
