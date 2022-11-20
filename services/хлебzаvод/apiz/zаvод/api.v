module main

import log
import vweb

struct ApiError {
	error string
}

fn (mut app App) error(code int, message string) vweb.Result {
	app.status = code.str()
	return app.json<ApiError>(ApiError{message})
}

fn (mut app App) internal_error(message string) vweb.Result {
	app.log.error('internal error: ${message}')
	app.status = '500'
	return app.text('Internal Server Error')
}
