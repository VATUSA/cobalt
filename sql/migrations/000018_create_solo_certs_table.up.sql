CREATE TABLE solo_cert (
    id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
    cid bigint NOT NULL,
    position varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_unicode_ci NOT NULL,
    expires date NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW(),
    CONSTRAINT solo_cert_cid_cid_fk
    FOREIGN KEY (cid) REFERENCES user(cid) ON DELETE CASCADE 
)