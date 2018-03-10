create table if not exists currency (
    currency_id int auto_increment primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table if not exists exchange_rate (
    currency_id           int            not null,
    exchange_date         date           not null,       
    rate                  numeric(18, 6) not null,
    reference_currency_id int            not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date),
    constraint exchange_rate_currency_fk foreign key (currency_id)
        references currency (currency_id),
    constraint exchange_rate_ref_currency_fk foreign key (reference_currency_id)
        references currency (currency_id)
);

create index idx_exchange_rate_date on exchange_rate (exchange_date);

create table if not exists audit_log (
    audit_log_id   bigint auto_increment PRIMARY KEY,
    source         varchar(64) not null,
    source_version varchar(16) not null,
    log_time       datetime(3) not null,
    log_msg        JSON not null
);

create index idx_time_audit_log on audit_log (log_time);
CREATE INDEX idx_log_source_audit_log ON audit_log (source);
