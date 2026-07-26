-- Describes tables owned and migrated by mithril (the v3 API), not cobalt.
-- This file exists only so sqlc can type-check queries against it; it is
-- never applied as a migration by cobalt. The source of truth for this
-- schema is mithril's sql/app_tables.sql.

CREATE TABLE `v3_api_key` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `code` VARCHAR(255) NOT NULL,
    `testing` BOOLEAN NOT NULL DEFAULT FALSE,
    `facility` VARCHAR(255) NULL,
    `notes` TEXT NULL,
    `created_at` BIGINT NOT NULL,
    `updated_at` BIGINT NULL,
    `deleted_at` BIGINT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `v3_api_key_code_unique` (`code`),
    KEY `v3_api_key_facility_index` (`facility`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
