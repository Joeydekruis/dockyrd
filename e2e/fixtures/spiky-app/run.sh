#!/bin/sh
# Allocates memory in spikes (simulates webpack-style peaks).
while true; do
  python3 - <<'PY'
import time
b = bytearray(200 * 1024 * 1024)
time.sleep(5)
del b
PY
  sleep 10
done
