CREATE SCHEMA IF NOT EXISTS wmeter AUTHORIZATION geo;

SET search_path TO public;

create or replace view dual as select 'X' AS dummy;

CREATE TABLE IF NOT EXISTS audit_log (
    audit_log_id   bigserial PRIMARY KEY,
    log_source     varchar(64) not null,
    log_time       timestamp not null,
    audit_msg      jsonb     not null
);

SET search_path TO wmeter,public;

CREATE TABLE IF NOT EXISTS system_params (
    system_params_id serial PRIMARY KEY,
    param_group      varchar(64) not null,
    param            varchar(64) not null,
    val              varchar(64) not null
);

CREATE UNIQUE INDEX IF NOT EXISTS system_params_uk
    ON system_params (param_group, param);

CREATE TABLE IF NOT EXISTS request (
    request_id        serial PRIMARY KEY,
    request_title     varchar(64)  not null DEFAULT '-',
    request_template  varchar(64)  not null DEFAULT '-',
    request_url       varchar(128) not null DEFAULT '-',
    controller        varchar(64)  not null DEFAULT '-',
    action            varchar(64)  not null DEFAULT '-',
    redirect_url      varchar(256) not null DEFAULT '-',
    redirect_on_error varchar(256) not null DEFAULT '-',
    request_type      varchar(8)   not null DEFAULT 'GET',
    constraint request_url_uk unique (request_url, request_type),
    constraint request_type_chk check (request_type in ('GET', 'POST'))
);

CREATE TABLE IF NOT EXISTS role (
    role_id   serial PRIMARY KEY,
    role      varchar(64) not null
);

CREATE UNIQUE INDEX IF NOT EXISTS role_uk ON role (lower(role));

CREATE TABLE IF NOT EXISTS "user" (
    user_id                bigserial PRIMARY KEY,
    username               varchar(64) not null,
    loweredusername        varchar(64) not null,
    name                   varchar(64) not null,
    surname                varchar(64) not null,
    email                  varchar(64) not null,
    loweredemail           varchar(64) not null,
    creation_time          timestamp   not null,
    last_update            timestamp   not null,
    activated              int         not null DEFAULT 0,
    activation_time        timestamp,
    last_password_change   timestamp,
    failed_password_atmpts int         not null DEFAULT 0,
    first_failed_password  timestamp,
    last_failed_password   timestamp,
    last_connect_time      timestamp,
    last_connect_ip        varchar(128),
    valid                  int         not null DEFAULT 1,
    locked_out             int         not null DEFAULT 0,
    constraint user_uk unique(loweredusername)
);

CREATE TABLE IF NOT EXISTS user_password (
    password_id   bigserial PRIMARY KEY,
    user_id       bigint       not null,
    password      varchar(256) not null,
    password_salt varchar(256) not null,
    valid_from    timestamp    not null,
    valid_until   timestamp,
    temporary     int          not null DEFAULT 0,
    constraint user_password_fk foreign key (user_id)
      references "user"(user_id)
);

CREATE TABLE IF NOT EXISTS user_role (
    user_role_id bigserial PRIMARY KEY,
    user_id      bigint not null,
    role_id      int not null,
    valid_from   timestamp not null,
    valid_until  timestamp,
    constraint  user_role_fk foreign key (role_id)
        references role (role_id),
    constraint user_role_usr_fk foreign key (user_id)
      references "user"(user_id)
);

CREATE TABLE IF NOT EXISTS user_role_history (
    user_role_id bigint PRIMARY KEY,
    user_id      bigint not null,
    role_id      int not null,
    valid_from   timestamp not null,
    valid_until  timestamp,
    constraint  user_role_history_fk foreign key (role_id)
        references role (role_id),
    constraint user_role_usr_fk foreign key (user_id)
      references "user"(user_id)
);

CREATE TABLE IF NOT EXISTS user_ip (
  user_ip_id bigserial PRIMARY KEY,
  user_id    bigint       NOT NULL,
  ip         varchar(256) NOT NULL,
  constraint user_ip_fk foreign key (user_id)
    references "user"(user_id)
);

CREATE TABLE IF NOT EXISTS cookie_encode_key (
    cookie_encode_key_id serial       PRIMARY KEY,
    encode_key           varchar(256) not null,
    valid_from           timestamp not null,
    valid_until          timestamp not null
);
