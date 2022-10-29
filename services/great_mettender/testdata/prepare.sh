#!/bin/bash -e

gzip -c "${1}"  | zstd -c | base64 > "${1}.gz.zst.b64"
