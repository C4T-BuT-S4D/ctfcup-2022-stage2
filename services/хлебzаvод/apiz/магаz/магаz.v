module main

import auth
import db
import graceful
import log
import os
import vweb

struct App {
	vweb.Context
	store db.Store     [vweb_global]
	auth  auth.Service [vweb_global]
mut:
	error string
	log   log.Log [vweb_global]
}

fn main() {
	graceful.register()!

	store := db.open(os.getenv('DB_HOST'), os.getenv('DB_USER'), os.getenv('DB_PASSWORD')) or {
		panic('failed to connect to db: ${err}')
	}

	service := auth.new_service(os.getenv('AUTH_HOST'))
	service.sign('test') or { panic('failed to connect to auth service: ${err}') }

	mut app := &App{
		log: log.Log{
			level: .debug
		}
		store: store
		auth: service
	}
	app.mount_static_folder_at('./public', '/public')

	vweb.run(app, 80)
}

['/order'; get]
pub fn (mut app App) order() vweb.Result {
	mut order := db.Order{}
	mut capsule := Capsule{}

	capsule = app.unpack_capsule(app.query['voucher']) or {
		app.error = 'У вас недействительный талон! Удостоверьтесь в том, что вы его правильно ввели.'
		$vweb.html()
		return app.ok('')
	}

	order = app.store.get_order(capsule.order_id) or {
		app.error = 'Хлеб ещё не испечён, придите позже!'
		$vweb.html()
		return app.ok('')
	}
	return $vweb.html()
}
