module graceful

import os
import time

const (
	shutdown_timeout = time.millisecond * 500
)

pub fn register() ! {
	os.signal_opt(.int, graceful_shutdown)!
	os.signal_opt(.term, graceful_shutdown)!
}

// Not really graceful, but ok
fn graceful_shutdown(_ os.Signal) {
	println('shutting down in ${shutdown_timeout}')
	time.sleep(shutdown_timeout)
	exit(0)
}
