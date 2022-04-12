## Migrations
Create new migration: `./mongo/migrate.sh create -ext json -seq -dir path/to/migrations/ migration_name`.
For example: `./mongo/migrate.sh create -ext json -seq -dir ./service-go/migrations/confa confa_init`

Run migrations: `make migrate`

Visit [gomigrate docs](https://github.com/golang-migrate/migrate) to learn more.
