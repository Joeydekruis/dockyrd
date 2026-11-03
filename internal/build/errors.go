package build

import "fmt"

func errUnsupportedBuilder(name string) error {
	return fmt.Errorf("unsupported builder %q", name)
}
