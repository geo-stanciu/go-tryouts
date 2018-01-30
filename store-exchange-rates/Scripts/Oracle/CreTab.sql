create sequence s$currency nocache start with 1;

create table currency (
    currency_id number default s$currency.nextval primary key,
    currency    varchar2(8) not null,
    constraint currency_uk unique (currency)
);

create table exchange_rate (
    currency_id           number        not null,
    exchange_date         date          not null,       
    rate                  number(18, 6) not null,
    reference_currency_id number        not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date),
    constraint exchange_rate_currency_fk foreign key (currency_id)
        references currency (currency_id),
    constraint exchange_rate_ref_currency_fk foreign key (reference_currency_id)
        references currency (currency_id)
);

create index idx_exchange_rate_date on exchange_rate (exchange_date);

create sequence s$audit_log nocache start with 1;

create table audit_log (
    audit_log_id   number default s$audit_log.nextval primary key,
    log_source     varchar2(64) not null,
    log_time       timestamp not null,
    audit_msg      long      not null
);

create index idx_time_audit_log on audit_log (log_time);
CREATE INDEX idx_log_source_audit_log ON audit_log (log_source);
