from html_checker import HTMLChecker, Tag
from typing import Tuple
import requests

PORT = 41385

class CheckMachine:
  @property
  def url(self):
    return f"http://{self.c.host}:{self.port}"

  def __init__(self, checker: HTMLChecker):
    self.c = checker
    self.port = PORT

  def ensure_recipient(self, session: requests.Session, link: str, want: str):
    self.c.assert_eq(link.startswith(f"{self.url}/order?voucher="), True, "Invalid order link")
    
    response = session.get(link)
    soup = self.c.get_soup(response, "Invalid order link")

    info_tag = soup.select_one('body>.section>.container>.box>.content>h3')
    self.c.assert_neq(info_tag, None, "No info found by order link")

    text = info_tag.get_text()
    needle = ', Ваш свежеиспечённый хлеб'
    self.c.assert_in(needle, text, "No info found by order link")
    
    have = text[:text.index(needle)]
    self.c.assert_eq(have, want, "No info found by order link")