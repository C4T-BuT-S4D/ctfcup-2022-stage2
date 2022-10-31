#!/usr/bin/env python

import subprocess
import sys
from pathlib import Path

BASE_DIR = Path(__file__).absolute().parent
BIN_PATH = BASE_DIR / 'gmt_checker'

p = subprocess.run([str(BIN_PATH), *sys.argv[1:]], shell=False, check=False)
sys.exit(p.returncode)
