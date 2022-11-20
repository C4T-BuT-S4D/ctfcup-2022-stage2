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
	log log.Log [vweb_global]
}

fn main() {
	store := db.open(os.getenv('DB_HOST'), os.getenv('DB_USER'), os.getenv('DB_PASSWORD')) or {
		panic('failed to connect to db: ${err}')
	}

	service := auth.new_service(os.getenv('AUTH_HOST'))
	service.sign('test') or { panic('failed to connect to auth service: ${err}') }

	app := &App{
		log: log.Log{
			level: .debug
		}
		store: store
		auth: service
	}

	vweb.run(app, 80)
}

['/api/receive/:voucher'; get]
pub fn (mut app App) receive_order(voucher string) vweb.Result {
	id := app.auth.unsign(voucher) or {
		app.status = '400'
		return app.text('У вас недействительный талон! Удостоверьтесь в том, что вы его правильно ввели.')
	}

	order := app.store.get_order(id) or {
		app.status = '418'
		return app.text('Хлеб ещё не испечён, придите позже!')
	}
	return app.json<db.Order>(order)
}
