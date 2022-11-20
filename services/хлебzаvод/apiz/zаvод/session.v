module main

import vweb

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
