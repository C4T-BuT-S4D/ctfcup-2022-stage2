module main

// import auth
import db

fn main() {
	// db := pg.connect(pg.Config{
	// 	host: 'localhost'
	// 	port: 5432
	// 	user: 'zavod'
	// 	password: 'zavod-password'
	// 	dbname: 'xlebzavod'
	// })!

	store := db.open('localhost', 'zavod', 'zavod-password') or {
		println('failed to connect to db: ${err}')
		return
	}

	store.authenticate_user('aboba', 'amoga') or {
		println(err)
		return
	}

	// id := store.create_order('aboba', 'tandyr') or {
	// 	println(err)
	// 	return
	// }
	// println(id)

	order := store.get_order('') or {
		if db.is_error(err, .not_found) {
			println('no such order')
			return
		}
		println(err)
		return
	}
	println(order)

	orders := store.list_orders('aboba')!
	println(orders)

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
	// // token := service.sign('aboba') or {
	// // 	println(err)
	// // 	return
	// // }
	// // println(token)

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
