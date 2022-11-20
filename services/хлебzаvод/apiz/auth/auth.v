module auth

import crypto.rc4
import encoding.base64
import json
import net.http

pub struct Service {
	url string
}

pub struct TokenError {
	Error
}

fn (err TokenError) msg() string {
	return 'invalid token'
}

pub fn new_service(host string) &Service {
	return &Service{
		url: 'http://${host}'
	}
}

pub fn (s &Service) sign(data string) !string {
	resp := http.post('${s.url}/sign', data)!
	if resp.status_code != 200 {
		return error('received bad status code ${resp.status_code}')
	}

	return resp.body
}

pub fn (s &Service) unsign(token string) !string {
	resp := http.get('${s.url}/unsign/${token}')!
	if resp.status_code != 200 {
		return IError(TokenError{})
	}

	parts := token.split('.')
	if parts.len < 1 || parts[0].len < 1 {
		return IError(TokenError{})
	}

	return decrypt(base64.url_decode_str(parts[0]), resp.body)
}

pub fn (s &Service) sign_json<T>(v T) !string {
	return s.sign(json.encode(v))
}

pub fn (s &Service) unsign_json<T>(token string) !T {
	data := s.unsign(token)!
	return json.decode(T, data)
}

fn decrypt(data string, key string) !string {
	mut c := rc4.new_cipher(key.bytes())!

	mut m := data.bytes()
	c.xor_key_stream(mut m, mut m)
	return m.bytestr()
}
