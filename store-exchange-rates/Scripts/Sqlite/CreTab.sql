create view if not exists dual as select 'X' AS dummy;

create table if not exists currency (
    currency_id integer primary key autoincrement,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table if not exists exchange_rate (
    currency_id           integer        not null,
    exchange_date         date           not null,
    rate                  numeric(18, 6) not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date),
    constraint exchange_rate_currency_fk foreign key (currency_id)
        references currency (currency_id)
);

create index if not exists idx_exchange_rate_curr_id on exchange_rate (currency_id);
create index if not exists idx_exchange_rate_date on exchange_rate (exchange_date);

create table if not exists audit_log (
    audit_log_id   integer PRIMARY KEY autoincrement,
    source         varchar(64) not null,
    source_version varchar(16) not null,
    log_time       datetime(3) not null,
    log_msg        text not null
);

create index if not exists idx_time_audit_log on audit_log (log_time);
create index if not exists idx_log_source_audit_log ON audit_log (source);
