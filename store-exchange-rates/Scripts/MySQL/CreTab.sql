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
)
partition by range (year(exchange_date)) (
    partition p2005 values less than (2006),
    partition p2006 values less than (2007),
    partition p2007 values less than (2008),
    partition p2008 values less than (2009),
    partition p2009 values less than (2010),
    partition p2010 values less than (2011),
    partition p2011 values less than (2012),
    partition p2012 values less than (2013),
    partition p2013 values less than (2014),
    partition p2014 values less than (2015),
    partition p2015 values less than (2016),
    partition p2016 values less than (2017),
    partition p2017 values less than (2018),
    partition p2018 values less than (2019),
    partition p2019 values less than (2020),
    partition p2020 values less than (2021),
    partition pmax values less than (MAXVALUE)
);

create table if not exists audit_log (
    audit_log_id   bigint auto_increment PRIMARY KEY,
    log_time       datetime not null DEFAULT current_timestamp,
    audit_msg      MEDIUMTEXT not null
)
partition by range (audit_log_id) (
    partition p1mil values less than (1000000),
    partition p2mil values less than (2000000),
    partition pmax values less than (MAXVALUE)
);

create index idx_time_audit_log on audit_log (log_time);
