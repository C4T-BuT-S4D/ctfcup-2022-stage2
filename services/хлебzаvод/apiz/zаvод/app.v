module main

import log
import net.http
import vweb

// This is needed because in V $vweb.html() does not actually return from the function...
// Jesus, f*cking, Christ...
fn (mut app App) ret() vweb.Result {
	return app.ok('')
}

fn (mut app App) error(status http.Status, message string) {
	app.status = status.int().str()
	app.error = message
}

fn (mut app App) internal_error(message string) vweb.Result {
	app.log.error('internal error: ${message}')
	app.status = '500'
	return app.text('Internal Server Error')
}
