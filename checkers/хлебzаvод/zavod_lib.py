from html_checker import HTMLChecker, Tag
from typing import Tuple
from checklib import Status
import requests
import re

PORT = 41384
order_id_re = re.compile(r"#(\d{10})")

class CheckMachine:
  @property
  def url(self):
    return f"http://{self.c.host}:{self.port}"

  def __init__(self, checker: HTMLChecker):
    self.c = checker
    self.port = PORT

  def validate_index(self, session: requests.Session, breads: list[str]):
    response = session.get(self.url)
    soup = self.c.get_soup(response, "Invalid auth page")
    self.c.assert_tags(
      "body>.section>.container>.box>figure>img",
      list(map(lambda b: Tag(name="img", attrs={"src": f"/public/img/{b}.jpg"}), breads)),
      soup,
      "Invalid index page",
    )
  
  def auth(self, session: requests.Session, action: str, username: str, password: str, status=Status.MUMBLE):
    response = session.get(f"{self.url}/{action}")
    soup = self.c.get_soup(response, f"Invalid {action} page", status=status)
    self.c.assert_tags(
      "body>.section>.container>.box input",
      [Tag(name="input", attrs={"name": "username"}), Tag(name="input", attrs={"name": "password"})],
      soup,
      f"Invalid {action} page",
      status=status
    )

    response = session.post(response.url, data={"username": username, "password": password})
    self.c.get_soup(response, f"Failed to {action}", status=status)
    self.c.assert_in("session", session.cookies, f"No session after {action}", status=status)
    self.c.assert_eq(response.url.rstrip("/"), self.url, f"Invalid page after {action}", status=status)

  def create_order(self, session: requests.Session, bread: str, recipient: str) -> Tuple[str, str]:
    response = session.get(f"{self.url}/order/{bread}")
    soup = self.c.get_soup(response, "Invalid order page")
    self.c.assert_tags(
      "body>.section>.container>.box input",
      [Tag(name="input", attrs={"name": "recipient"})],
      soup,
      "Invalid order page"
    )

    response = session.post(response.url, data={"recipient": recipient})
    soup = self.c.get_soup(response, "Failed to place order")
    
    order_id_tag = soup.select_one('body>.section>.container>.box>.content>h3')
    self.c.assert_neq(order_id_tag, None, "Order ID not found on placed order")
    order_id_match = order_id_re.search(order_id_tag.string or '')
    self.c.assert_neq(order_id_match, None, "Order ID not found on placed order")
    order_id = order_id_match.group(1)

    order_link_tag = soup.select_one('body>.section>.container>.box>.content>p>a')
    self.c.assert_neq(order_id_tag, None, "Order link not found on placed order")
    self.c.assert_in("href", order_link_tag.attrs, "Order link not found on placed order")
    order_link = order_link_tag.attrs["href"]

    return (order_id, order_link)

  def ensure_orders(self, session: requests.Session, orders: list[Tuple[str, str]], status=Status.MUMBLE):
    response = session.get(f"{self.url}/orders")
    soup = self.c.get_soup(response, "Invalid orders page", status=status)
    self.c.assert_tags(
      "body>.section>.container>.box>.title",
      list(map(lambda order: Tag(name="div", value=f"#{order[1]}"), orders)),
      soup,
      f"Order not found on orders page {session.cookies}",
      status=status
    )
    self.c.assert_tags(
      "body>.section>.container>.box>figure>img",
      list(map(lambda order: Tag(name="img", attrs={"src": f"/public/img/{order[0]}.jpg"}), orders)),
      soup,
      f"Order not found on orders page {session.cookies}",
      status=status
    )
    