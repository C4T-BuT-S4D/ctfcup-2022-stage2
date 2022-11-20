module main

import auth
import vweb

const (
	auth_prefix = 'Bearer '
)

struct Session {
	username string
}

struct AuthResponse {
	token string
}

fn (mut app App) authorize(username string) vweb.Result {
	token := app.auth.sign_json<Session>(Session{username}) or {
		return app.internal_error('signing session using auth service: ${err}')
	}
	return app.json<AuthResponse>(AuthResponse{token})
}

pub fn (mut app App) before_request() {
	mut token := app.req.header.get(.authorization) or { return }

	if !token.starts_with(auth_prefix) {
		return
	}
	token = token.after(auth_prefix)

	session := app.auth.unsign_json<Session>(token) or {
		if err is auth.TokenError {
			return
		}

		app.error(503, 'failed to auth: ${err}')
		return
	}

	app.session = session
}
