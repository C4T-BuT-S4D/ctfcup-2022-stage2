module auth

pub struct TokenError {
	Error
}

fn (err TokenError) msg() string {
	return 'invalid token'
}
