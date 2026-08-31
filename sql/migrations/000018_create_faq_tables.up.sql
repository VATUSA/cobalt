CREATE TABLE faq_category (
    id integer not null auto_increment primary key,
    title varchar(120) not null,
    sort_order integer not null default 0
);

CREATE TABLE faq_item (
    id integer not null auto_increment primary key,
    faq_category_id integer not null,
    question varchar(500) not null,
    answer text not null,
    sort_order integer not null default 0,
    FOREIGN KEY (faq_category_id) REFERENCES faq_category(id) ON DELETE RESTRICT
);
