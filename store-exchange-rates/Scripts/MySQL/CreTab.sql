create table if not exists currency (
    currency_id int auto_increment primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table if not exists exchange_rate (
    currency_id      int            not null,
    exchange_date    date           not null,       
    rate             numeric(18, 6) not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date)
);

create table if not exists audit_log (
    audit_log_id   bigint auto_increment PRIMARY KEY,
    log_source     varchar(64) not null,
    log_time       datetime(3) not null,
    audit_msg      MEDIUMTEXT not null
);

create index idx_time_audit_log on audit_log (log_time);
