/*
download timezone_2018e_posix_sql.zip - POSIX standard
from https://dev.mysql.com/downloads/timezones.html
copy the sql file in your prefered directory

cd to your directory

run:

mysql -u root -p mysql

then run

MySQL [mysql]> source timezone_posix.sql;

*/

CREATE DATABASE devel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER 'geo'@'%' identified by "geo";

use devel;

GRANT ALL PRIVILEGES ON devel to 'geo'@'%';

GRANT ALL PRIVILEGES ON devel.* to 'geo'@'%';

FLUSH PRIVILEGES;
