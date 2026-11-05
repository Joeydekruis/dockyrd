package resources

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func readMemTotal() (int64, error) {
	return readMeminfoKey("MemTotal")
}

func readSwapFree() (int64, error) {
	return readMeminfoKey("SwapFree")
}

func readMeminfoKey(key string) (int64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, key) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0, err
		}
		return kb * 1024, nil
	}
	return 0, os.ErrNotExist
}
