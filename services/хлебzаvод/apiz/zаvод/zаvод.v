module main

import auth
import db
import graceful
import log
import os
import vweb

struct App {
	vweb.Context
	store      db.Store     [vweb_global]
	auth       auth.Service [vweb_global]
	magaz_port string       [vweb_global]
mut:
	session Session
	error   string
	log     log.Log [vweb_global]
}

fn main() {
	graceful.register()!

	store := db.open(os.getenv('DB_HOST'), os.getenv('DB_USER'), os.getenv('DB_PASSWORD')) or {
		panic('failed to connect to db: ${err}')
	}

	service := auth.new_service(os.getenv('AUTH_HOST'))
	service.sign_json<Session>(Session{'test-session'}) or {
		panic('failed to connect to auth service: ${err}')
	}

	mut app := &App{
		log: log.Log{
			level: .debug
		}
		store: store
		auth: service
		magaz_port: os.getenv('MAGAZ_PORT')
	}
	app.mount_static_folder_at('./public', '/public')

	vweb.run(app, 80)
}
