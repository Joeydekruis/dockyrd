//go:build !unix

package watch

// sameFilesystem on non-Unix dev hosts compares paths only.
func sameFilesystem(a, b string) (bool, error) {
	return a == b, nil
}
