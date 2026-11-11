package hostops

import (
	"context"
	"strings"
)

// SSHPasswordAuthEnabled reports effective PasswordAuthentication for a typical root login.
func SSHPasswordAuthEnabled(ctx context.Context, exec Commander) (bool, error) {
	out, err := exec.Run(ctx, "sshd", "-T", "-C", "user=root,host=localhost,addr=127.0.0.1")
	if err != nil {
		out, err = exec.Run(ctx, "sshd", "-T")
		if err != nil {
			return false, err
		}
	}
	return parseSSHDPasswordAuth(string(out)), nil
}

func parseSSHDPasswordAuth(cfg string) bool {
	last := ""
	for _, line := range strings.Split(strings.ToLower(cfg), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "passwordauthentication ") {
			continue
		}
		last = strings.TrimSpace(strings.TrimPrefix(line, "passwordauthentication "))
	}
	return last == "yes"
}
