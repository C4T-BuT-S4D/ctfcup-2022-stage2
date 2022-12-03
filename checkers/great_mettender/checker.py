#!/usr/bin/env python3
import os
import subprocess
import sys
from pathlib import Path

BASE_DIR = Path(__file__).absolute().parent
BIN_PATH = BASE_DIR / 'gmt_checker'

env = os.environ
env['QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING'] = 'true'
p = subprocess.run([str(BIN_PATH), *sys.argv[1:]], shell=False, check=False, env=env)
sys.exit(p.returncode)
