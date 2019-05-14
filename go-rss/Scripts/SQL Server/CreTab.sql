create view dual as select 'X' AS dummy;

create table rss_source (
    rss_source_id       int identity(1,1) primary key not null,
    source_name         varchar(256) not null,
    lowered_source_name varchar(256) not null,
    language            varchar(8) not null,
    copyright           nvarchar(max),
    source_link         nvarchar(max),
    title               nvarchar(max),
    description         nvarchar(max),
    last_rss_date       datetime2(3) not null DEFAULT '1970-01-01 00:00:00',
    add_date            datetime2(3) not null,
    generator           nvarchar(max),
    web_master          nvarchar(max),
    image_title         nvarchar(max),
    image_width         nvarchar(max),
    image_heigth        nvarchar(max),
    image_link          nvarchar(max),
    image_url           nvarchar(max),
    constraint rss_source_uk unique (lowered_source_name)
);

create table rss (
    rss_id             bigint identity(1,1) primary key not null,
    rss_source_id      int not null,
    title              nvarchar(max) not null,
    link               nvarchar(max),
    description        nvarchar(max),
    item_guid          nvarchar(max),
    orig_link          nvarchar(max),
    rss_date           datetime2(3) not null,
    add_date           datetime2(3) not null,
    keywords           nvarchar(max),
    category           nvarchar(max),
    subcategory        nvarchar(max),
    content            nvarchar(max),
    tags               nvarchar(max),
    creator            nvarchar(max),
    enclosure_link     nvarchar(max),
    enclosure_length   int,
    enclosure_filetype nvarchar(max),
    media_link         nvarchar(max),
    media_filetype     nvarchar(max),
    media_thumbnail    nvarchar(max),
    seen               int not null default 0,
    constraint rss_source_fk foreign key (rss_source_id)
        references rss_source (rss_source_id)
);

create index idx_rss_source_id on rss (rss_source_id);
create index idx_rss_date on rss (rss_date);
create index idx_rss_item on rss(rss_id) include (title, link);
create index idx_rss_seen on rss(seen);

create table audit_log (
    audit_log_id   bigint identity(1,1) PRIMARY KEY,
    source         varchar(64) not null,
    source_version varchar(16) not null,
    log_time       datetime2(3) not null,
    log_msg        nvarchar(max) not null
);

create index idx_time_audit_log on audit_log (log_time);
create index idx_log_source_audit_log ON audit_log (source);
