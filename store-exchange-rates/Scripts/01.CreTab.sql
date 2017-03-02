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
