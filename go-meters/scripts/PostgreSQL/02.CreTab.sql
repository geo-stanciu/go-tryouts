CREATE TABLE wmeter.system_params (
    system_params_id serial PRIMARY KEY,
    param_group      varchar(64) not null,
    param            varchar(64) not null,
    val              varchar(64) not null
);

CREATE UNIQUE INDEX system_params_uk
    ON wmeter.system_params (lower(param_group), lower(param));

CREATE TABLE wmeter.request (
    request_id        serial PRIMARY KEY,
    request_title     varchar(64)  not null DEFAULT '-',
    request_template  varchar(64)  not null DEFAULT '-',
    request_url       varchar(128) not null DEFAULT '-',
    controller        varchar(64)  not null DEFAULT '-',
    action            varchar(64)  not null DEFAULT '-',
    redirect_url      varchar(256) not null DEFAULT '-',
    redirect_on_error varchar(256) not null DEFAULT '-',
    constraint request_url_uk unique (request_url)
);

CREATE TABLE wmeter.role (
    role_id   serial PRIMARY KEY,
    role      varchar(64) not null
);

CREATE UNIQUE INDEX role_uk ON wmeter.role (lower(role));

CREATE TABLE wmeter.user (
    user_id                serial PRIMARY KEY,
    username               varchar(64) not null,
    loweredusername        varchar(64) not null,
    name                   varchar(64) not null,
    surname                varchar(64) not null,
    email                  varchar(64) not null,
    loweredemail           varchar(64) not null,
    creation_time          timestamp with time zone not null DEFAULT current_timestamp,
    last_update            timestamp with time zone not null DEFAULT current_timestamp,
    activated              int         not null DEFAULT 0,
    activation_time        timestamp with time zone,
    last_password_change   timestamp with time zone,
    failed_password_atmpts int         not null DEFAULT 0,
    first_failed_password  timestamp with time zone,
    last_failed_password   timestamp with time zone,
    last_connect_time      timestamp with time zone,
    last_connect_ip        varchar(128),
    valid                  int         not null DEFAULT 1,
    locked_out             int         not null DEFAULT 0
);

CREATE TABLE wmeter.user_password (
    password_id   serial PRIMARY KEY,
    user_id       int          not null,
    password      varchar(256) not null,
    password_salt varchar(256) not null,
    valid_from    timestamp with time zone not null DEFAULT current_timestamp,
    valid_until   timestamp with time zone,
    temporary     int          not null DEFAULT 0
);

CREATE TABLE wmeter.user_role (
    user_role_id serial PRIMARY KEY,
    user_id      int not null,
    role_id      int not null,
    valid_from   timestamp with time zone not null DEFAULT current_timestamp,
    valid_until  timestamp with time zone,
    constraint  user_role_fk foreign key (role_id)
        references wmeter.role (role_id)
);

CREATE TABLE wmeter.user_role_history (
    user_role_id int PRIMARY KEY,
    user_id      int not null,
    role_id      int not null,
    valid_from   timestamp with time zone not null,
    valid_until  timestamp with time zone,
    constraint  user_role_history_fk foreign key (role_id)
        references wmeter.role (role_id)
);

CREATE TABLE wmeter.audit_log (
    audit_log_id   bigserial PRIMARY KEY,
    log_time       timestamp not null DEFAULT (current_timestamp at time zone 'UTC'),
    audit_msg      jsonb       not null
);

CREATE TABLE wmeter.cookie_encode_key (
    cookie_encode_key_id serial       PRIMARY KEY,
    encode_key           varchar(256) not null,
    valid_from           timestamp with time zone not null DEFAULT statement_timestamp(),
    valid_until          timestamp with time zone not null DEFAULT statement_timestamp() + interval '30' day
);
