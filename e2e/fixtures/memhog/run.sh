#!/bin/sh
# Allocates ~400MB RSS then holds (for build OOM / budget tests on small VMs).
MB="${1:-400}"
exec python3 -c "import time; b=bytearray(int('${MB}')*1024*1024); print('allocated', len(b));
while True: time.sleep(60)"
