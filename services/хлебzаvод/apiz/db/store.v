module db

import pg

pub enum ErrorCode as u8 {
	not_found
	duplicate
}

pub struct StoreError {
	Error
	e ErrorCode
}

fn (err StoreError) msg() string {
	return 'store error: ${err.e}'
}

pub fn is_error(err IError, code ErrorCode) bool {
	if err is StoreError {
		return err.e == code
	}
	return false
}

struct Store {
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
