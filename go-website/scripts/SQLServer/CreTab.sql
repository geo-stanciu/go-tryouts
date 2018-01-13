create view dual as select 'X' AS dummy;

create table currency (
    currency_id int identity(1,1) primary key,
    currency    varchar(8) not null,
    constraint currency_uk unique (currency)
);

create table exchange_rate (
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

create table audit_log (
    audit_log_id   bigint identity(1,1) PRIMARY KEY,
    log_source     varchar(64) not null,
    log_time       datetime not null,
    audit_msg      ntext not null
);

create index idx_time_audit_log on audit_log (log_time);
CREATE INDEX idx_log_source_audit_log ON audit_log (log_source);


CREATE TABLE system_params (
  system_params_id int identity(1,1) PRIMARY KEY,
  param_group      varchar(64) not null,
  param            varchar(64) not null,
  val              varchar(64) not null
);

CREATE UNIQUE INDEX system_params_uk
  ON system_params (param_group, param);

CREATE TABLE request (
  request_id        int identity(1,1) PRIMARY KEY,
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

CREATE TABLE role (
  role_id   int identity(1,1) PRIMARY KEY,
  role      varchar(64) not null
);

CREATE UNIQUE INDEX role_uk ON role (role);

CREATE TABLE "user" (
  user_id                bigint identity(1,1) PRIMARY KEY,
  username               nvarchar(64) not null,
  loweredusername        nvarchar(64) not null,
  name                   nvarchar(64) not null,
  surname                nvarchar(64) not null,
  email                  nvarchar(64) not null,
  loweredemail           nvarchar(64) not null,
  creation_time          datetime not null,
  last_update            datetime not null,
  activated              int         not null DEFAULT 0,
  activation_time        datetime,
  last_password_change   datetime,
  failed_password_atmpts int         not null DEFAULT 0,
  first_failed_password  datetime,
  last_failed_password   datetime,
  last_connect_time      datetime,
  last_connect_ip        varchar(128),
  valid                  int         not null DEFAULT 1,
  locked_out             int         not null DEFAULT 0,
  CONSTRAINT user_uk unique(loweredusername)
);

CREATE TABLE user_password (
  password_id   bigint       identity(1,1) PRIMARY KEY,
  user_id       bigint       NOT NULL,
  password      VARCHAR(256) NOT NULL,
  password_salt VARCHAR(256) NOT NULL,
  valid_from    datetime NOT NULL,
  valid_until   datetime,
  temporary     INT          NOT NULL DEFAULT 0,
  constraint user_password_fk foreign key (user_id)
    references "user"(user_id)
);

CREATE TABLE user_role (
  user_role_id bigint identity(1,1) PRIMARY KEY,
  user_id      bigint not null,
  role_id      int not null,
  valid_from   datetime not null,
  valid_until  datetime,
  constraint user_role_fk foreign key (role_id)
    references role(role_id),
  constraint user_role_usr_fk foreign key (user_id)
    references "user"(user_id)
);

CREATE TABLE user_role_history (
  user_role_id bigint PRIMARY KEY,
  user_id      bigint not null,
  role_id      int not null,
  valid_from   datetime not null,
  valid_until  datetime,
  constraint user_role_h_fk foreign key (role_id)
    references role(role_id),
  constraint user_role_h_usr_fk foreign key (user_id)
    references "user"(user_id)
);

CREATE TABLE user_ip (
  user_ip_id bigint       identity(1,1) PRIMARY KEY,
  user_id    bigint       NOT NULL,
  ip         varchar(256) NOT NULL,
  constraint user_ip_fk foreign key (user_id)
    references "user"(user_id)
);

CREATE TABLE cookie_encode_key (
  cookie_encode_key_id int identity(1,1) PRIMARY KEY,
  encode_key           varchar(256) not null,
  valid_from           datetime not null,
  valid_until          datetime not null
);