CREATE TABLE system_params (
  system_params_id integer PRIMARY KEY AUTOINCREMENT,
  param_group      varchar(64) not null,
  param            varchar(64) not null,
  val              varchar(64) not null
);

CREATE UNIQUE INDEX system_params_uk
  ON system_params (param_group, param);

CREATE TABLE request (
  request_id        integer PRIMARY KEY AUTOINCREMENT,
  request_title     varchar(64)  not null DEFAULT '-',
  request_template  varchar(64)  not null DEFAULT '-',
  request_url       varchar(128) not null DEFAULT '-',
  controller        varchar(64)  not null DEFAULT '-',
  action            varchar(64)  not null DEFAULT '-',
  redirect_url      varchar(256) not null DEFAULT '-',
  redirect_on_error varchar(256) not null DEFAULT '-',
  constraint request_url_uk unique (request_url)
);

CREATE TABLE role (
  role_id   integer PRIMARY KEY AUTOINCREMENT,
  role      varchar(64) not null
);

CREATE UNIQUE INDEX role_uk ON role (role);

CREATE TABLE user (
  user_id                integer PRIMARY KEY AUTOINCREMENT,
  username               varchar(64) not null,
  loweredusername        varchar(64) not null,
  name                   varchar(64) not null,
  surname                varchar(64) not null,
  email                  varchar(64) not null,
  loweredemail           varchar(64) not null,
  creation_time          timestamp not null DEFAULT current_timestamp,
  last_update            timestamp not null DEFAULT current_timestamp,
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
  CONSTRAINT user_uk unique(loweredusername)
);

CREATE TABLE user_password (
  password_id   integer PRIMARY KEY AUTOINCREMENT,
  user_id       INT          NOT NULL,
  password      VARCHAR(256) NOT NULL,
  password_salt VARCHAR(256) NOT NULL,
  valid_from    timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  valid_until   timestamp,
  temporary     INT          NOT NULL DEFAULT 0,
  constraint user_password_fk foreign key (user_id)
      references user(user_id)
);

CREATE TABLE user_role (
  user_role_id integer PRIMARY KEY AUTOINCREMENT,
  user_id      int not null,
  role_id      int not null,
  valid_from   timestamp not null DEFAULT current_timestamp,
  valid_until  timestamp,
  constraint user_role_fk foreign key (user_id)
      references user(user_id),
  constraint user_role_role_fk foreign key (role_id)
      references role(role_id)
);

CREATE TABLE user_role_history (
  user_role_id int PRIMARY KEY,
  user_id      int not null,
  role_id      int not null,
  valid_from   timestamp not null,
  valid_until  timestamp,
  constraint user_role_history_fk foreign key (user_id)
      references user(user_id),
  constraint user_role_history_role_fk foreign key (role_id)
      references role(role_id)
);

create table audit_log (
  audit_log_id   integer PRIMARY KEY autoincrement,
  log_time       timestamp not null DEFAULT current_timestamp,
  audit_msg      text not null
);

CREATE TABLE cookie_encode_key (
  cookie_encode_key_id integer PRIMARY KEY autoincrement,
  encode_key           varchar(256) not null,
  valid_from           timestamp not null DEFAULT current_timestamp,
  valid_until          timestamp not null
);
