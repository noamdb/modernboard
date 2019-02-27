package repository

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type Repository struct {
	db *sqlx.DB
}

const pageSize = 30

func (d *Repository) Connect(dataSourceName string) {
	con, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Panic(err)
	}
	d.db = con
	d.db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
}
