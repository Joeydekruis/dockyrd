//go:build !unix

package diskusage

func sameFilesystem(a, b string) (bool, error) {
	return a == b, nil
}
