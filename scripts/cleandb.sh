DB_NAME=${1:-ebakus}

echo "* Clean postgres db"
psql postgres -c "drop database $DB_NAME"

echo "\n* Create postgres db"
createdb $DB_NAME

echo "\n* Create db structure from schema"
psql -d $DB_NAME -a -f $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/schema/create.sql

echo "\n* Clean redis"
redis-cli FLUSHDB
