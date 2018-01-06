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

create sequence s$audit_log nocache start with 1;

create table audit_log (
    audit_log_id   number default s$audit_log.nextval primary key,
    log_source     varchar2(64) not null,
    log_time       timestamp not null,
    audit_msg      long      not null
);

create index idx_time_audit_log on audit_log (log_time);
CREATE INDEX idx_log_source_audit_log ON audit_log (log_source);


create sequence s$system_params nocache start with 1;

CREATE TABLE system_params (
    system_params_id number default s$system_params.nextval PRIMARY KEY,
    param_group      varchar2(64) not null,
    param            varchar2(64) not null,
    val              varchar2(64) not null
);

CREATE UNIQUE INDEX system_params_uk
    ON system_params (param_group, param);

create sequence s$request nocache start with 1;

CREATE TABLE request (
    request_id        number default s$request.nextval PRIMARY KEY,
    request_title     varchar2(64)  DEFAULT '-' not null,
    request_template  varchar2(64)  DEFAULT '-' not null,
    request_url       varchar2(128) DEFAULT '-' not null,
    controller        varchar2(64)  DEFAULT '-' not null,
    action            varchar2(64)  DEFAULT '-' not null,
    redirect_url      varchar2(256) DEFAULT '-' not null,
    redirect_on_error varchar2(256) DEFAULT '-' not null,
    request_type      varchar2(8)   DEFAULT 'GET' not null,
    constraint request_url_uk unique (request_url, request_type),
    constraint request_type_chk check (request_type in ('GET', 'POST'))
);

create sequence s$role nocache start with 1;

CREATE TABLE role (
    role_id   number default s$role.nextval PRIMARY KEY,
    role      varchar2(64) not null
);

CREATE UNIQUE INDEX role_uk ON role (role);

create sequence s$user nocache start with 1;

CREATE TABLE "user" (
    user_id                number default s$user.nextval PRIMARY KEY,
    username               nvarchar2(128) not null,
    loweredusername        nvarchar2(128) not null,
    name                   nvarchar2(128) not null,
    surname                nvarchar2(128) not null,
    email                  nvarchar2(128) not null,
    loweredemail           nvarchar2(128) not null,
    creation_time          timestamp   not null,
    last_update            timestamp   not null,
    activated              number      DEFAULT 0 not null,
    activation_time        timestamp,
    last_password_change   timestamp,
    failed_password_atmpts number      DEFAULT 0 not null,
    first_failed_password  timestamp,
    last_failed_password   timestamp,
    last_connect_time      timestamp,
    last_connect_ip        varchar2(128),
    valid                  number      DEFAULT 1 not null,
    locked_out             number      DEFAULT 0 not null,
    constraint user_uk unique(loweredusername)
);

create sequence s$user_password nocache start with 1;

CREATE TABLE user_password (
    password_id   number default s$user_password.nextval PRIMARY KEY,
    user_id       number        not null,
    password      varchar2(256) not null,
    password_salt varchar2(256) not null,
    valid_from    timestamp     not null,
    valid_until   timestamp,
    temporary     number        DEFAULT 0 not null,
    constraint user_password_fk foreign key (user_id)
      references "user"(user_id)
);

create sequence s$user_role nocache start with 1;

CREATE TABLE user_role (
    user_role_id number default s$user_role.nextval PRIMARY KEY,
    user_id      number not null,
    role_id      number not null,
    valid_from   timestamp not null,
    valid_until  timestamp,
    constraint  user_role_fk foreign key (role_id)
        references role (role_id),
    constraint user_role_usr_fk foreign key (user_id)
      references "user"(user_id)
);

CREATE TABLE user_role_history (
    user_role_id number PRIMARY KEY,
    user_id      number not null,
    role_id      number not null,
    valid_from   timestamp not null,
    valid_until  timestamp,
    constraint  user_role_history_fk foreign key (role_id)
        references role (role_id),
    constraint user_role_h_usr_fk foreign key (user_id)
      references "user"(user_id)
);

create sequence s$user_ip nocache start with 1;

CREATE TABLE user_ip (
  user_ip_id number default s$user_ip.nextval PRIMARY KEY,
  user_id    number       NOT NULL,
  ip         varchar2(256) NOT NULL,
  constraint user_ip_fk foreign key (user_id)
    references "user"(user_id)
);

create sequence s$cookie_encode_key nocache start with 1;

CREATE TABLE cookie_encode_key (
    cookie_encode_key_id number default s$user_ip.nextval PRIMARY KEY,
    encode_key           varchar2(256) not null,
    valid_from           timestamp not null,
    valid_until          timestamp not null
);
