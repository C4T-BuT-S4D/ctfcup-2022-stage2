module main

import db
import vweb

const (
	bread_kinds = ['white', 'wheat', 'grain', 'rye', 'bagel', 'baguette', 'pita', 'ciabatta',
		'focaccia']
)

struct CreateOrderResponse {
	voucher string
}

struct ListOrdersResponse {
	orders []db.Order
}

['/api/orders'; post]
pub fn (mut app App) create_order(bread string, recipient string) vweb.Result {
	if app.session.username == '' {
		return app.error(401, 'unauthorized')
	} else if bread == '' {
		return app.error(422, 'необходимо указать вид хлеба')
	} else if recipient == '' {
		return app.error(422, 'необходимо указать получателя заказа')
	} else if bread !in bread_kinds {
		return app.error(422, 'указан недоступный тип хлеба')
	}

	id := app.store.create_order(app.session.username, bread, recipient) or {
		return app.internal_error('creating order: ${err}')
	}
	return app.json<CreateOrderResponse>(CreateOrderResponse{id})
}

['/api/orders'; get]
pub fn (mut app App) list_orders() vweb.Result {
	if app.session.username == '' {
		return app.error(401, 'unauthorized')
	}

	orders := app.store.list_orders(app.session.username) or {
		return app.internal_error('listing orders: ${err}')
	}
	return app.json<ListOrdersResponse>(ListOrdersResponse{orders})
}
