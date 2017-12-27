CREATE TABLE system_params (
  system_params_id int AUTO_INCREMENT PRIMARY KEY,
  param_group      varchar(64) not null,
  param            varchar(64) not null,
  val              varchar(64) not null
);

CREATE UNIQUE INDEX system_params_uk
  ON system_params (param_group, param);

CREATE TABLE request (
  request_id        int AUTO_INCREMENT PRIMARY KEY,
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
  role_id   int AUTO_INCREMENT PRIMARY KEY,
  role      varchar(64) not null
);

CREATE UNIQUE INDEX role_uk ON role (role);

CREATE TABLE user (
  user_id                bigint AUTO_INCREMENT PRIMARY KEY,
  username               varchar(64) not null,
  loweredusername        varchar(64) not null,
  name                   varchar(64) not null,
  surname                varchar(64) not null,
  email                  varchar(64) not null,
  loweredemail           varchar(64) not null,
  creation_time          datetime(3) not null,
  last_update            datetime(3) not null,
  activated              int         not null DEFAULT 0,
  activation_time        datetime(3),
  last_password_change   datetime(3),
  failed_password_atmpts int         not null DEFAULT 0,
  first_failed_password  datetime(3),
  last_failed_password   datetime(3),
  last_connect_time      datetime(3),
  last_connect_ip        varchar(128),
  valid                  int         not null DEFAULT 1,
  locked_out             int         not null DEFAULT 0,
  CONSTRAINT user_uk unique(loweredusername)
);

CREATE TABLE user_password (
  password_id   bigint       AUTO_INCREMENT PRIMARY KEY,
  user_id       bigint       NOT NULL,
  password      VARCHAR(256) NOT NULL,
  password_salt VARCHAR(256) NOT NULL,
  valid_from    datetime(3) NOT NULL,
  valid_until   datetime(3),
  temporary     INT          NOT NULL DEFAULT 0,
  constraint user_password_fk foreign key (user_id)
    references user(user_id)
);

CREATE TABLE user_role (
  user_role_id bigint            AUTO_INCREMENT PRIMARY KEY,
  user_id      bigint not null,
  role_id      int not null,
  valid_from   datetime(3) not null,
  valid_until  datetime(3),
  constraint user_role_fk foreign key (role_id)
    references role(role_id),
  constraint user_role_usr_fk foreign key (user_id)
    references user(user_id)
);

CREATE TABLE user_role_history (
  user_role_id bigint PRIMARY KEY,
  user_id      bigint not null,
  role_id      int not null,
  valid_from   datetime(3) not null,
  valid_until  datetime(3),
  constraint user_role_h_fk foreign key (role_id)
    references role(role_id),
  constraint user_role_h_usr_fk foreign key (user_id)
    references user(user_id)
);

CREATE TABLE user_ip (
  user_ip_id bigint       AUTO_INCREMENT PRIMARY KEY,
  user_id    bigint       NOT NULL,
  ip         varchar(256) NOT NULL,
  constraint user_ip_fk foreign key (user_id)
    references user(user_id)
);

create table audit_log (
  audit_log_id   bigint auto_increment PRIMARY KEY,
  log_source     varchar(64) not null,
  log_time       datetime(3) not null,
  audit_msg      JSON not null
);

CREATE TABLE cookie_encode_key (
  cookie_encode_key_id int auto_increment PRIMARY KEY,
  encode_key           varchar(256) not null,
  valid_from           datetime(3) not null,
  valid_until          datetime(3) not null
);
