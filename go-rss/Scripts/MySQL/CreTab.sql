create table if not exists rss_source (
    rss_source_id       int auto_increment primary key not null,
    source_name         varchar(256) not null,
    lowered_source_name varchar(256) not null,
    language            varchar(8) not null,
    copyright           text,
    source_link         text,
    title               text,
    description         text,
    last_rss_date       datetime(3) not null DEFAULT '1970-01-01 00:00:00',
    add_date            datetime(3) not null,
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
    rss_id             bigint auto_increment primary key not null,
    rss_source_id      int not null,
    title              text not null,
    link               text,
    description        mediumtext,
    item_guid          text,
    orig_link          text,
    rss_date           datetime(3) not null,
    add_date           datetime(3) not null,
    keywords           text,
    category           text,
    subcategory        text,
    content            mediumtext,
    tags               text,
    creator            text,
    enclosure_link     text,
    enclosure_length   int,
    enclosure_filetype text,
    media_link         text,
    media_filetype     text,
    media_thumbnail    text,
    constraint rss_source_fk foreign key (rss_source_id)
        references rss_source (rss_source_id)
);

create index if not exists idx_rss_source_id on rss (rss_source_id);
create index if not exists idx_rss_date on rss (rss_date);
create index if not exists idx_rss_item on rss (title(256), link(256));

create table if not exists audit_log (
    audit_log_id   bigint auto_increment primary key not null,
    source         varchar(64) not null,
    source_version varchar(16) not null,
    log_time       datetime(3) not null,
    log_msg        JSON not null
);

create index if not exists idx_time_audit_log on audit_log (log_time);
create index if not exists idx_log_source_audit_log ON audit_log (source);
