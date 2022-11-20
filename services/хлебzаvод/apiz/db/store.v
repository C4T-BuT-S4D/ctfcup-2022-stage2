module db

import pg

pub struct Store {
	db pg.DB
}

pub fn open(host string, username string, password string) !&Store {
	db := pg.connect(pg.Config{
		host: host
		port: 5432
		user: username
		password: password
		dbname: 'xlebzavod'
	})!

	return &Store{
		db: db
	}
}
