## Quick Start

Fetch the project with the proper go tool.

`$ go get -d -v bitbucket.org/pantelisss/ebakus_server`

Build both executables.

`$ $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/scripts/build.sh`

Connect to your database and execute the provided sql file located in the schema directory to create all the required tables.

`$ psql -d YOUR_DB_NAME -a -f ./schema/create.sql`

Start the ebakus node. For example:

`$ ebakus-linux-amd64 --testnet`

To use the crawler run:

`$ $GOPATH/bin/ebakus_crawler fetchblocks --ipc ~/.ebakus/testnet/ebakus.ipc --dbname YOUR_DB_NAME --dbuser YOUR_DB_USER --dbpass YOUR_DB_PASS`

To start the explorer (webapi) run:

`$ $GOPATH/bin/ebakus_explorer`
