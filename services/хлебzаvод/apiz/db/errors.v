module db

pub enum ErrorCode as u8 {
	not_found
	duplicate
}

pub struct StoreError {
	Error
	e ErrorCode
}

fn (err StoreError) msg() string {
	return 'store error: ${err.e}'
}

fn known_error(code ErrorCode) IError {
	return IError(StoreError{
		e: code
	})
}

pub fn is_error(err IError, code ErrorCode) bool {
	if err is StoreError {
		return err.e == code
	}
	return false
}
