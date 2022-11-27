module main

import auth
import time
import net.http
import vweb

const (
	auth_cookie          = 'session'
	cookie_lifetime_days = 1
)

struct Session {
	username string
}

fn (mut app App) authorize(username string) vweb.Result {
	token := app.auth.sign_json<Session>(Session{username}) or {
		return app.internal_error('signing session using auth service: ${err}')
	}

	app.set_cookie(http.Cookie{
		name: auth_cookie
		value: token
		expires: time.now().add_days(cookie_lifetime_days)
		http_only: true
	})
	return app.redirect('/')
}

pub fn (mut app App) before_request() {
	token := app.get_cookie(auth_cookie) or { return }
	session := app.auth.unsign_json<Session>(token) or {
		if err is auth.TokenError {
			return
		}

		app.error(.service_unavailable, 'failed to auth: ${err}')
		return
	}

	app.session = session
}
