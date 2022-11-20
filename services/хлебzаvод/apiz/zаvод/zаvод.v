module main

import auth
import db
import log
import os
import vweb

struct App {
	vweb.Context
	store db.Store     [vweb_global]
	auth  auth.Service [vweb_global]
mut:
	session Session
	log     log.Log [vweb_global]
}

fn main() {
	store := db.open(os.getenv('DB_HOST'), os.getenv('DB_USER'), os.getenv('DB_PASSWORD')) or {
		panic('failed to connect to db: ${err}')
	}

	service := auth.new_service(os.getenv('AUTH_HOST'))
	service.sign_json<Session>(Session{'test-session'}) or {
		panic('failed to connect to auth service: ${err}')
	}

	app := &App{
		log: log.Log{
			level: .debug
		}
		store: store
		auth: service
	}

	vweb.run(app, 80)
}
