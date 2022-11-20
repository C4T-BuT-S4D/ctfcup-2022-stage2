module db

pub fn (s &Store) create_user(username string, password string) ! {
	s.db.exec_param2("insert into users values ($1, crypt($2, gen_salt('bf')))", username,
		password) or {
		if err.msg().contains('duplicate key value violates unique constraint "users_pkey"') {
			return IError(StoreError{
				e: .duplicate
			})
		}
		return err
	}
}

pub fn (s &Store) authenticate_user(username string, password string) ! {
	row := s.db.exec_param2('select 1 from users where username = $1 and password = crypt($2, password)',
		username, password)!

	if row.len != 1 {
		return IError(StoreError{
			e: .not_found
		})
	}
}
