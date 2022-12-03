module main

import auth
import hash.crc32
import time

const (
	err_bad_capsule = error('bad capsule')
	order_id_len    = 10
)

struct Capsule {
	order_id  string
	timestamp time.Time
}

pub fn (mut app App) unpack_capsule(voucher string) !Capsule {
	capsule := app.auth.unsign(voucher) or {
		if err !is auth.TokenError {
			app.log.error('failed to unsign voucher using auth service: ${err}')
		}
		return err_bad_capsule
	}

	parts := capsule.split('|')
	if parts.len != 3 || parts[1].len != order_id_len {
		return err_bad_capsule
	}

	unsigned := '${parts[0]}|${parts[1]}'
	crc := crc32.sum(unsigned.bytes())
	if crc.hex_full() != parts[2] {
		return err_bad_capsule
	}

	ts := time.parse(parts[0]) or { return err_bad_capsule }
	return Capsule{
		order_id: parts[1]
		timestamp: ts
	}
}
