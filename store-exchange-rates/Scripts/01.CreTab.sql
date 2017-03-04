create table if not exists currency (
    currency_id serial primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table if not exists exchange_rate (
    exchange_rate_id serial primary key,
    currency_id      int            not null,
    exchange_date    date           not null,       
    rate             numeric(18, 6) not null,
    constraint exchange_rate_currency_fk foreign key (currency_id)
        references currency (currency_id)
);

create table if not exists audit_log (
    audit_log_id   bigserial PRIMARY KEY,
    log_time       timestamp not null DEFAULT statement_timestamp(),
    audit_msg      jsonb     not null
);
