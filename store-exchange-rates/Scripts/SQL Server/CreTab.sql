create view dual as select 'X' AS dummy;

create table currency (
    currency_id int identity(1,1) primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table exchange_rate (
    currency_id      int            not null,
    exchange_date    date           not null,       
    rate             numeric(18, 6) not null,
    constraint exchange_rate_pk primary key (currency_id, exchange_date)
);

create table audit_log (
    audit_log_id   bigint identity(1,1) PRIMARY KEY,
    log_source     varchar(64) not null,
    log_time       datetime not null,
    audit_msg      text not null
);

create index idx_time_audit_log on audit_log (log_time);
