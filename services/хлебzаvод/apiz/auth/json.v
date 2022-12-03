module auth

import json

pub fn (s &Service) sign_json<T>(v T) !string {
	return s.sign(json.encode(v))
}

pub fn (s &Service) unsign_json<T>(token string) !T {
	data := s.unsign(token)!
	return json.decode(T, data)
}
