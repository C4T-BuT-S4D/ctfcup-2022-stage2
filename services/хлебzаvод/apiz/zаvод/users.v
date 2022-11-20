module main

import db
import regex
import vweb

['/api/user'; post]
pub fn (mut app App) create_user(username string, password string) vweb.Result {
	if username.len < 3 {
		return app.error(422, 'ник должен быть не короче 3-х символов')
	} else if password.len < 5 {
		return app.error(422, 'пароль должен быть не короче 5-ти символов')
	} else if username.len > 10 || password.len > 10 {
		return app.error(422, 'ник и пароль не должны быть длиннее 10-ти символов')
	}

	mut username_re := regex.regex_opt(r'^[a-z\d]{3,10}$') or {
		return app.internal_error('compiling username regex: ${err}')
	}
	if !username_re.matches_string(username) {
		return app.error(422, 'ник может состоять только из строчных букв латинского алфавита и цифр')
	}

	app.store.create_user(username, password) or {
		if db.is_error(err, .duplicate) {
			return app.error(409, 'пользователь с таким ником уже зарегистрирован')
		}
		return app.internal_error('creating user: ${err}')
	}

	return app.authorize(username)
}

['/api/session'; post]
pub fn (mut app App) create_session(username string, password string) vweb.Result {
	app.store.authenticate_user(username, password) or {
		if db.is_error(err, .not_found) {
			return app.error(401, 'не существует пользователя с таким ником или паролем')
		}
		return app.internal_error('authenticating user: ${err}')
	}

	return app.authorize(username)
}
