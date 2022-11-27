module main

import db
import regex
import vweb

['/register'; get; post]
pub fn (mut app App) register() vweb.Result {
	if app.session.username != '' {
		return app.redirect('/')
	}

	if app.req.method == .get {
		$vweb.html()
		return app.ret()
	}

	username := app.form['username']
	password := app.form['password']
	validated := app.register_validate(username, password) or {
		return app.internal_error('${err}')
	}
	if !validated {
		$vweb.html()
		return app.ret()
	}

	app.store.create_user(username, password) or {
		if db.is_error(err, .duplicate) {
			app.error(.conflict, 'Пользователь с таким ником уже зарегистрирован')
			$vweb.html()
			return app.ret()
		}
		return app.internal_error('creating user: ${err}')
	}

	return app.authorize(username)
}

fn (mut app App) register_validate(username string, password string) !bool {
	if username.len < 3 {
		app.error(.unprocessable_entity, 'Ник должен быть не короче 3-х символов')
	} else if password.len < 5 {
		app.error(.unprocessable_entity, 'Пароль должен быть не короче 5-ти символов')
	} else if username.len > 10 || password.len > 10 {
		app.error(.unprocessable_entity, 'Ник и пароль не должны быть длиннее 10-ти символов')
	} else {
		mut username_re := regex.regex_opt(r'^[a-z\d]{3,10}$') or {
			return error('compiling username regex: ${err}')
		}
		
		if !username_re.matches_string(username) {
			app.error(.unprocessable_entity, 'Ник может состоять только из строчных букв латинского алфавита и цифр')
			return false
		}
		return true
	}
	return false
}

['/login'; get; post]
pub fn (mut app App) login() vweb.Result {
	if app.session.username != '' {
		return app.redirect('/')
	}

	if app.req.method == .get {
		$vweb.html()
		return app.ret()
	}

	username := app.form['username']
	password := app.form['password']
	app.store.authenticate_user(username, password) or {
		if db.is_error(err, .not_found) {
			app.error(.unauthorized, 'Не существует пользователя с таким ником или паролем')
			$vweb.html()
			return app.ret()
		}
		return app.internal_error('authenticating user: ${err}')
	}

	return app.authorize(username)
}
