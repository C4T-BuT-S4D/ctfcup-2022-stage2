from bs4 import Tag as bs4Tag, BeautifulSoup
import checklib
import mimeparse

class Tag(bs4Tag):
  value: str | None

  def __init__(self, value=None, **kwargs):
    super().__init__(**kwargs)
    self.value = value

class HTMLChecker(checklib.BaseChecker):
  def get_soup(self, r, public: str, status=checklib.status.Status.MUMBLE) -> BeautifulSoup:
    self.assert_eq(r.status_code, 200, public, status)
    self.assert_in("Content-Type", r.headers, public, status)
    try:
      mimetype, subtype, _ = mimeparse.parse_mime_type(r.headers["Content-Type"])
      self.assert_eq(mimetype, "text", public, status)
      self.assert_eq(subtype, "html", public, status)
    except mimeparse.MimeTypeParseException:
      self.cquit(status, public, f'{r.url} returned invalid Content-Type')

    r.encoding = 'utf-8'
    data = self.get_text(r, public, status)

    # html.parser never throws exceptions, but this shouldn't hurt
    try:
      soup = BeautifulSoup(data, "html.parser")
    except Exception:
      self.cquit(status, public, f'Failed to parse html from {r.url}')
    return soup
  
  def assert_tags(self, selector: str, tags: list[Tag], soup: BeautifulSoup, public: str, status=checklib.status.Status.MUMBLE):
    selected = soup.select(selector)
    self.assert_eq(len(selected), len(tags), public, status)
    for have, want in zip(selected, tags):
      self.assert_eq(have.name, want.name, public, status)
      for key, value in want.attrs.items():
        self.assert_in(key, have.attrs, public, status)
        self.assert_eq(have.attrs[key], value, public, status)

      if want.value is not None:
        self.assert_eq(have.get_text(), want.value, public, status)