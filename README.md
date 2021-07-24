## Migrations
Create new migration: `./mongo/migrate.sh create -ext json -seq -dir path/to/migrations/ migration_name`. Path to migrations is considered relative to the `internal` directory.
For example: `./mongo/migrate.sh create -ext json -seq -dir ./confa/migrations confa_init`

Run migrations: `make migrate`

Visit [gomigrate docs](https://github.com/golang-migrate/migrate) to learn more.
