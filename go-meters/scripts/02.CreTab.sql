CREATE TABLE wmeter.page (
    page_id       serial PRIMARY KEY,
    page_title    varchar(64)  not null,
    page_template varchar(64)  not null,
    controller    varchar(64)  not null,
    action        varchar(64)  not null,
    page_url      varchar(256) not null,
    constraint page_url_uk unique (page_url)
);

CREATE TABLE wmeter.request (
    request_id        serial PRIMARY KEY,
    request_url       varchar(256) not null,
    controller        varchar(64)  not null,
    action            varchar(64)  not null,
    redirect_url      varchar(256) not null,
    redirect_on_error varchar(256) not null
    constraint request_url_uk unique (request_url)
);

CREATE TABLE wmeter.user (
    user_id         serial PRIMARY KEY,
    username        varchar(64) not null,
    name            varchar(64) not null,
    surname         varchar(64) not null,
    email           varchar(64) not null,
    loweredusername varchar(64) not null,
    loweredemail    varchar(64) not null,
    valid           int,
    constraint lower_user_uk unique (loweredusername),
    constraint lower_email_uk unique (loweredemail)
);

CREATE UNIQUE INDEX username_uk
    ON wmeter.user (lower(username));

CREATE TABLE wmeter.user_password (
    password_id   serial PRIMARY KEY,
    user_id       int          not null,
    password      varchar(256) not null,
    password_salt varchar(256) not null,
    valid_from    timestamp    not null DEFAULT statement_timestamp(),
    valid_until   timestamp,
    constraint user_password_fk foreign key (user_id)
        references wmeter.user(user_id)
);

CREATE TABLE wmeter.user_password_archive (
    password_id   int PRIMARY KEY,
    user_id       int          not null,
    password      varchar(256) not null,
    password_salt varchar(256) not null,
    valid_from    timestamp    not null,
    valid_until   timestamp,
    constraint user_password_archive_fk foreign key (user_id)
        references wmeter.user(user_id)
);

CREATE TABLE wmeter.audit_log (
    audit_log_id   bigserial PRIMARY KEY,
    log_time       timestamp   not null DEFAULT statement_timestamp(),
    audit_msg      jsonb       not null
);
