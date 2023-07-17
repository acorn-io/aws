#!/bin/sh
set -e

(
    sleep 60
    kill $PPID
) &

while ! mysql --connect-timeout=2 -h ${MYSQL_HOST} -u ${MYSQL_ADMIN_USER} -p${MYSQL_ADMIN_PASSWORD} -e "SELECT 1"; do
    sleep .5
done

echo Creating database $MYSQL_DATABASE and user $MYSQL_USER with user $MYSQL_ADMIN_USER
mysql -h ${MYSQL_HOST} -u ${MYSQL_ADMIN_USER} -p${MYSQL_ADMIN_PASSWORD} << EOF
CREATE DATABASE IF NOT EXISTS ${MYSQL_DATABASE};
CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'%';
ALTER USER '${MYSQL_USER}'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}';
GRANT ALL PRIVILEGES ON ${MYSQL_DATABASE}.* TO '${MYSQL_USER}'@'%';
EOF
