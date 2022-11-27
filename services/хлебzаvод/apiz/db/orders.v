module db

import arrays
import pg

pub struct Order {
pub:
	id        string
	bread     string
	recipient string
}

pub fn (s &Store) create_order(username string, bread string, recipient string) !string {
	rows := s.db.exec_param_many('insert into orders (username, bread, recipient) values ($1, $2, $3) returning id',
		[username, bread, recipient])!
	return rows[0].vals[0]
}

pub fn (s &Store) list_orders(username string) ![]Order {
	rows := s.db.exec_param('select id, bread from orders where username = $1', username)!
	return arrays.map_indexed<pg.Row, Order>(rows, fn (_ int, row pg.Row) Order {
		return Order{
			id: row.vals[0]
			bread: row.vals[1]
		}
	})
}

pub fn (s &Store) get_order(id string) !Order {
	rows := s.db.exec_param('select bread, recipient from orders where id = $1', id)!
	if rows.len != 1 {
		return known_error(.not_found)
	}

	vals := rows[0].vals
	return Order{
		bread: vals[0]
		recipient: vals[1]
	}
}
