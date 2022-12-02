#!/usr/bin/env python3

from html_checker import HTMLChecker
from checklib import *
import json
import magaz_lib
import requests
import secrets
import string
import sys
import zavod_lib

BREADS = ["white", "wheat", "grain", "rye", "bagel", "baguette", "pita", "ciabatta", "focaccia"]
MAX_USERNAME_LEN = 15
MAX_PASSWORD_LEN = 15

class Checker(HTMLChecker):
  vulns: int = 1
  timeout: int = 15
  uses_attack_data: bool = True

  def __init__(self, *args, **kwargs):
    super(Checker, self).__init__(*args, **kwargs)
    self.zcm = zavod_lib.CheckMachine(self)
    self.mcm = magaz_lib.CheckMachine(self)

  def action(self, action, *args, **kwargs):
    try:
      super(Checker, self).action(action, *args, **kwargs)
    except requests.exceptions.ConnectionError:
      self.cquit(Status.DOWN, 'Connection error', 'Got requests connection error')

  def random_creds(self):
    username = rnd_username(MAX_USERNAME_LEN)[:MAX_USERNAME_LEN]
    password = rnd_password(MAX_PASSWORD_LEN)
    return username, password

  def check(self):
    session = get_initialized_session()
    self.zcm.validate_index(session, BREADS)

    username, password = self.random_creds()
    self.zcm.auth(session, "register", username, password)

    bread = secrets.choice(BREADS)
    recipient = rnd_username()
    order_id, order_link = self.zcm.create_order(session, bread, recipient)

    self.zcm.ensure_orders(session, [(bread, order_id)])
    self.mcm.ensure_recipient(get_initialized_session(), order_link, recipient)

    self.cquit(Status.OK)

  def put(self, _: str, flag: str, __: str):
    session = get_initialized_session()

    username, password = self.random_creds()
    self.zcm.auth(session, "register", username, password)

    bread = secrets.choice(BREADS)
    order_id, order_link = self.zcm.create_order(session, bread, flag)

    self.zcm.ensure_orders(session, [(bread, order_id)])
    self.cquit(Status.OK, order_id, json.dumps({
      "username": username,
      "password": password,
      "bread": bread,
      "order_id": order_id,
      "order_link": order_link,
    }))

  def get(self, flag_id: str, flag: str, _: str):
    session = get_initialized_session()
    data = json.loads(flag_id)

    username, password = data["username"], data["password"]
    self.zcm.auth(session, "login", username, password)

    bread = data["bread"]
    order_id, order_link = data["order_id"], data["order_link"]
    self.zcm.ensure_orders(session, [(bread, order_id)])

    self.mcm.ensure_recipient(get_initialized_session(), order_link, flag)
    self.cquit(Status.OK)

if __name__ == '__main__':
    c = Checker(sys.argv[2])

    try:
        c.action(sys.argv[1], *sys.argv[3:])
    except c.get_check_finished_exception():
        cquit(Status(c.status), c.public, c.private)
