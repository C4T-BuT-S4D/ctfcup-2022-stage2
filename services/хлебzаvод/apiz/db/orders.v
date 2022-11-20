module db

import time

pub struct Order {
	bread     string
	recipient string
	ts        time.Time
}

pub fn (s &Store) create_order(username string, bread string, recipient string) !string {
	rows := s.db.exec_param_many('insert into orders (username, bread, recipient) values ($1, $2, $3) returning id',
		[username, bread, recipient])!
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
	rows := s.db.exec_param('select bread, recipient, created_at from orders where id = $1',
		id) or {
		if err.msg().contains('invalid input syntax for type uuid') {
			return known_error(.not_found)
		}
		return err
	}

	if rows.len != 1 {
		return known_error(.not_found)
	}

	vals := rows[0].vals
	return Order{
		bread: vals[0]
		recipient: vals[1]
		ts: time.parse(vals[2])!
	}
}
