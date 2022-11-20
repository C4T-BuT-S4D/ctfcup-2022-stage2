module db

import time

struct Order {
	bread string
	ts    time.Time
}

pub fn (s &Store) create_order(username string, bread string) !string {
	rows := s.db.exec_param2('insert into orders (username, bread) values ($1, $2) returning id',
		username, bread)!
	return rows[0].vals[0]
}

pub fn (s &Store) list_orders(username string) ![]Order {
	rows := s.db.exec_param('select bread, created_at from orders where username = $1',
		username)!
	mut orders := []Order{len: rows.len}

	for i, row in rows {
		orders[i] = Order{
			bread: row.vals[0]
			ts: time.parse(row.vals[1])!
		}
	}
	return orders
}

pub fn (s &Store) get_order(id string) !Order {
	rows := s.db.exec_param('select bread, created_at from orders where id = $1', id) or {
		if err.msg().contains('invalid input syntax for type uuid') {
			return IError(StoreError{
				e: .not_found
			})
		}
		return err
	}

	if rows.len != 1 {
		return IError(StoreError{
			e: .not_found
		})
	}

	return Order{
		bread: rows[0].vals[0]
		ts: time.parse(rows[0].vals[1])!
	}
}
