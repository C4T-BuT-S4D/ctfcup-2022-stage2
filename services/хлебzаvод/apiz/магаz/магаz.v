module main

import auth
import db
import log
import os
import time
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

fn (mut app App) internal_error(message string) vweb.Result {
	app.log.error('internal error: ${message}')
	app.status = '500'
	return app.text('Internal Server Error')
}

struct Order {
	ts        time.Time
	bread     string
	recipient string
}

['/api/receive/:voucher'; get]
pub fn (mut app App) receive_order(voucher string) vweb.Result {
	timestamped := app.auth.unsign(voucher) or {
		app.status = '400'
		return app.text('У вас недействительный талон! Удостоверьтесь в том, что вы его правильно ввели.')
	}

	parts := timestamped.split('|')
	if parts.len != 2 {
		return app.internal_error('bad voucher received')
	}

	id := parts[1]
	ts := time.parse(parts[0]) or {
		return app.internal_error('unable to parse voucher time: ${err}')
	}

	order := app.store.get_order(id) or {
		app.status = '418'
		return app.text('Хлеб ещё не испечён, придите позже!')
	}
	return app.json<Order>(Order{ ts: ts, bread: order.bread, recipient: order.recipient })
}
