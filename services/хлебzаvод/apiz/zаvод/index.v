module main

import vweb

['/'; get]
pub fn (mut app App) index() vweb.Result {
	return $vweb.html()
}
