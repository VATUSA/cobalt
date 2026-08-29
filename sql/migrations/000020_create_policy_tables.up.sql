CREATE TABLE policy_category (
    id integer not null auto_increment primary key,
    title varchar(120) not null,
    sort_order integer not null default 0
);

CREATE TABLE policy_document (
    id integer not null auto_increment primary key,
    policy_category_id integer not null,
    ident varchar(20) not null,
    title varchar(255) not null,
    summary varchar(500) not null default '',
    document_url varchar(500) not null,
    effective_date date not null,
    hidden boolean not null default false,
    sort_order integer not null default 0,
    created_by_cid int not null,
    updated_by_cid int not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    FOREIGN KEY (policy_category_id) REFERENCES policy_category(id) ON DELETE RESTRICT
);
