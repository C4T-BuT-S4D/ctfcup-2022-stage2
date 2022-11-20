module main

import auth

struct Kek {
	name string
	n    int
}

fn main() {
	service := auth.new_service('localhost')
	// token := service.sign('aboba') or {
	// 	println(err)
	// 	return
	// }
	// println(token)

	// result := service.unsign(token) or {
	// 	println('err: ${err}')
	// 	return
	// }
	// println(result)

	token := service.sign_json(Kek{ name: 'aboba', n: 1337 }) or {
		println(err)
		return
	}
	println(token)
	result := service.unsign_json<Kek>('Mp5AyRFOhWwDtksHRAFNhUcOQlVHTA./../Mp5AyRFOhWwDtksHRAFNhUcOQlVGAki8nw.c342wRs0bDb8aAH4PS7MXg.7xGVkFjsGEuGkK94.goKzMkLuDycppaI9mL8GEw') or {
		println('err: ${err}')
		return
	}
	println(result)
}
