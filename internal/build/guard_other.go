//go:build !unix

package build

func diskFreeRatio(path string) (free, total uint64, err error) {
	return 1, 1, nil
}
