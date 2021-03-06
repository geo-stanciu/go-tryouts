create or replace view dual as select 'X' AS dummy;

create table if not exists rss_source (
    rss_source_id       serial primary key not null,
    source_name         text not null,
    lowered_source_name text not null,
    language            varchar(8) not null,
    copyright           text,
    source_link         text,
    title               text,
    description         text,
    last_rss_date       timestamp not null DEFAULT '1970-01-01 00:00:00',
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

create index if not exists idx_rss_source_add_date on rss_source (add_date);

create table if not exists rss (
    rss_id             bigserial primary key not null,
    rss_source_id      int not null,
    title              text not null,
    link               text,
    description        text,
    item_guid          text,
    orig_link          text,
    rss_date           timestamp not null,
    add_date           timestamp not null,
    keywords           text,
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
    media_thumbnail    text,
    seen               int not null default 0,
    constraint rss_source_fk foreign key (rss_source_id)
        references rss_source (rss_source_id)
);

create index if not exists idx_rss_source_id on rss (rss_source_id);
create index if not exists idx_rss_date on rss (rss_date);
create index if not exists idx_rss_add_date on rss (add_date);
create index if not exists idx_rss_item on rss (title, link);
create index if not exists idx_rss_seen on rss (seen);

create table if not exists audit_log (
    audit_log_id   bigserial primary key,
    source         varchar(64) not null,
    source_version varchar(64) not null,
    log_time       timestamp not null,
    log_msg        jsonb     not null
);

create index if not exists idx_time_audit_log on audit_log (log_time);
create index if not exists idx_log_source_audit_log ON audit_log (source);
