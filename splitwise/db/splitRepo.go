package db

import "split/postgres"

type Repo struct {
	postgres.PostgresRepo
}

var DB Repo
