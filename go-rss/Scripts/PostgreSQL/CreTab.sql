create or replace view dual as select 'X' AS dummy;

create table if not exists rss_source (
    rss_source_id       serial primary key not null,
    source_name         text not null,
    lowered_source_name text not null,
    language            varchar(8) not null,
    copyright           text,
    source_link         text not null,
    title               text,
    description         text,
    last_rss_date       timestamp,
    add_date            timestamp not null,
    generator           text,
    web_master          text,
    image_title         text,
    image_width         text,
    image_heigth        text,
    image_link          text,
    image_url           text,
    constraint rss_source_uk unique (lowered_source_name)
);

create table if not exists rss (
    rss_id             bigserial primary key not null,
    rss_source_id      int not null,
    title              text not null,
    link               text,
    description        text,
    item_guid          text,
    rss_date           timestamp not null,
    add_date           timestamp not null,
    category           text,
    subcategory        text,
    content            text,
    tags               text,
    creator            text,
    enclosure_link     text,
    enclosure_length   int,
    enclosure_filetype text,
    media_link         text,
    media_filetype     text,
    constraint rss_source_fk foreign key (rss_source_id)
        references rss_source (rss_source_id)
);

create table if not exists audit_log (
    audit_log_id   bigserial primary key,
    source         varchar(64) not null,
    source_version varchar(64) not null,
    log_time       timestamp not null,
    log_msg        jsonb     not null
);

create index idx_time_audit_log on audit_log (log_time);
CREATE INDEX idx_log_source_audit_log ON audit_log (source);
