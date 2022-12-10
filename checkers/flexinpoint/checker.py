#!/usr/bin/env python3
import signal
import subprocess
import sys
from pathlib import Path

BASE_DIR = Path(__file__).absolute().parent
BIN_PATH = BASE_DIR / 'flexinpoint_checker'

p = subprocess.Popen([str(BIN_PATH), *sys.argv[1:]], shell=False)


def exit_handler(*_args):
    print('Killing subprocess', file=sys.stderr)
    p.kill()


signal.signal(signal.SIGINT, exit_handler)
signal.signal(signal.SIGTERM, exit_handler)

sys.exit(p.wait())
