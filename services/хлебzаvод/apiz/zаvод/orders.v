module main

import hash.crc32
import time
import vweb

const (
	max_order_id_len = 10
)

['/order/:bread'; get; post]
pub fn (mut app App) order(bread string) vweb.Result {
	if app.session.username == '' {
		return app.redirect('/login')
	} else if bread !in bread_kinds {
		return app.redirect('/')
	}

	mut ordered := false
	mut order_id := ''
	mut voucher := ''

	if app.req.method == .get {
		$vweb.html()
		return app.ret()
	}

	recipient := app.form['recipient']
	if !app.order_validate(bread, recipient) {
		$vweb.html()
		return app.ret()
	}

	order_id = app.store.create_order(app.session.username, bread, recipient) or {
		return app.internal_error('creating order: ${err}')
	}
	order_id = app.format_order_id(order_id)

	voucher = app.order_voucher(order_id) or {
		return app.internal_error('signing order id: ${err}')
	}
	voucher = 'http://${app.get_header('Host').before(':')}:${app.magaz_port}/order?voucher=${voucher}'

	ordered = true
	return $vweb.html()
}

fn (mut app App) order_validate(bread string, recipient string) bool {
	if recipient == '' {
		app.error(.unprocessable_entity, 'Необходимо указать получателя заказа')
	} else if bread !in bread_kinds {
		app.error(.unprocessable_entity, 'Указан недоступный тип хлеба')
	} else {
		return true
	}
	return false
}

fn (mut app App) format_order_id(order_id string) string {
	if order_id.len < max_order_id_len {
		return '0'.repeat(max_order_id_len - order_id.len) + order_id
	}
	return order_id
}

fn (mut app App) order_voucher(order_id string) !string {
	unsigned := '${time.now().format_ss_micro()}|${order_id}'
	crc := crc32.sum(unsigned.bytes())
	capsule := '${unsigned}|${crc.hex_full()}'

	return app.auth.sign(capsule)
}

['/orders'; get]
pub fn (mut app App) orders() vweb.Result {
	if app.session.username == '' {
		return app.redirect('/login')
	}

	orders := app.store.list_orders(app.session.username) or {
		return app.internal_error('listing orders: ${err}')
	}
	return $vweb.html()
}
