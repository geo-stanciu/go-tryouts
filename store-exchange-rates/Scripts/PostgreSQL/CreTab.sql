create or replace view dual as select 'X' AS dummy;

create table if not exists currency (
    currency_id serial primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table if not exists exchange_rate (
    currency_id      int            not null,
    exchange_date    date           not null,       
    rate             numeric(18, 6) not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date),
    constraint exchange_rate_currency_fk foreign key (currency_id)
        references currency (currency_id)
);

create table if not exists audit_log (
    audit_log_id   bigserial primary key,
    log_source     varchar(64) not null,
    log_time       timestamp not null,
    audit_msg      jsonb     not null
);

create index idx_time_audit_log on audit_log (log_time);
