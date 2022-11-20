module main

import auth
import db
import log
import vweb
import os

struct App {
	vweb.Context
	store   db.Store     [vweb_global]
	auth    auth.Service [vweb_global]
mut:
	session Session
	log log.Log [vweb_global]
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
			output_label: 'zаvод'
		}
		store: store
		auth: service
	}

	vweb.run(app, 80)

	// db := pg.connect(pg.Config{
	// 	host: 'localhost'
	// 	port: 5432
	// 	user: 'zavod'
	// 	password: 'zavod-password'
	// 	dbname: 'xlebzavod'
	// })

	// store.authenticate_user('aboba', 'amoga') or {
	// 	println(err)
	// 	return
	// }

	// // id := store.create_order('aboba', 'tandyr') or {
	// // 	println(err)
	// // 	return
	// // }
	// // println(id)

	// order := store.get_order('') or {
	// 	if db.is_error(err, .not_found) {
	// 		println('no such order')
	// 		return
	// 	}
	// 	println(err)
	// 	return
	// }
	// println(order)

	// orders := store.list_orders('aboba')!
	// println(orders)

	// row := db.exec_param2('select username from users where username = $1 and password = crypt($2, password)',
	// 	'aboba', 'amoga') or {
	// 	println(err)
	// 	return
	// }
	// println(row)
	// rows := db.exec_param2("insert into users values ($1, crypt($2, gen_salt('bf')))",
	// 	'aboba', 'amoga') or {
	// 	if err.msg().contains('duplicate key value violates unique constraint') {
	// 		println('duplicate user')
	// 		return
	// 	}
	// 	println(err)
	// 	return
	// }
	// println(rows)

	// service := auth.new_service('localhost')

	// // result := service.unsign(token) or {
	// // 	println('err: ${err}')
	// // 	return
	// // }
	// // println(result)

	// token := service.sign_json(Kek{ name: 'aboba', n: 1337 }) or {
	// 	println(err)
	// 	return
	// }
	// println(token)
	// result := service.unsign_json<Kek>('Mp5AyRFOhWwDtksHRAFNhUcOQlVHTA./../Mp5AyRFOhWwDtksHRAFNhUcOQlVGAki8nw.c342wRs0bDb8aAH4PS7MXg.7xGVkFjsGEuGkK94.goKzMkLuDycppaI9mL8GEw') or {
	// 	println('err: ${err}')
	// 	return
	// }
	// println(result)
}

['/api/user'; post]
pub fn (mut app App) create_user(username string, password string) vweb.Result {
	if username.len < 3 {
		return app.error(422, 'ник должен быть не короче 3-х символов')
	} else if password.len < 5 {
		return app.error(422, 'пароль должен быть не короче 5-ти символов')
	} else if username.len > 10 || password.len > 10 {
		return app.error(422, 'ник и пароль не должны быть длиннее 10-ти символов')
	}

	app.store.create_user(username, password) or {
		if db.is_error(err, .duplicate) {
			return app.error(409, 'пользователь с таким ником уже зарегистрирован')
		}
		return app.internal_error('creating user: ${err}')
	}

	return app.authorize(username)
}

['/api/session'; post]
pub fn (mut app App) create_session(username string, password string) vweb.Result {
	app.store.authenticate_user(username, password) or {
		if db.is_error(err, .not_found) {
			return app.error(401, 'не существует пользователя с таким ником или паролем')
		}
		return app.internal_error('authenticating user: ${err}')
	}

	return app.authorize(username)
}
