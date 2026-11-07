package hostops

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func sshPrecheck(ctx context.Context, exec Commander, skip bool) error {
	if skip {
		return nil
	}
	out, err := exec.Run(ctx, "sshd", "-T")
	if err != nil {
		return fmt.Errorf("%w: cannot read sshd config: %v", ErrPrecondition, err)
	}
	cfg := strings.ToLower(string(out))
	if !strings.Contains(cfg, "passwordauthentication yes") {
		return nil
	}
	if !hasPrivilegedAuthorizedKeys() {
		return fmt.Errorf("%w: password authentication is on but no authorized_keys for root or sudo users", ErrPrecondition)
	}
	if err := rejectPasswordOnlySSHSession(); err != nil {
		return err
	}
	return nil
}

func hasPrivilegedAuthorizedKeys() bool {
	if keyFileNonEmpty("/root/.ssh/authorized_keys") {
		return true
	}
	sudoUsers := sudoGroupUsers()
	for _, u := range sudoUsers {
		usr, err := user.Lookup(u)
		if err != nil {
			continue
		}
		if keyFileNonEmpty(filepath.Join(usr.HomeDir, ".ssh", "authorized_keys")) {
			return true
		}
	}
	return false
}

func sudoGroupUsers() []string {
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}
		if fields[0] != "sudo" && fields[0] != "wheel" {
			continue
		}
		members := strings.TrimSpace(fields[3])
		if members == "" {
			return nil
		}
		return strings.Split(members, ",")
	}
	return nil
}

func keyFileNonEmpty(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}

// rejectPasswordOnlySSHSession blocks when the current SSH login user has no authorized_keys (password-only risk).
func rejectPasswordOnlySSHSession() error {
	if strings.TrimSpace(os.Getenv("SSH_CONNECTION")) == "" {
		return nil
	}
	cur, err := user.Current()
	if err != nil {
		return fmt.Errorf("%w: cannot determine current user for SSH session check", ErrPrecondition)
	}
	ak := filepath.Join(cur.HomeDir, ".ssh", "authorized_keys")
	if keyFileNonEmpty(ak) {
		return nil
	}
	return fmt.Errorf("%w: active SSH session for %q without authorized_keys (use key-based login before disabling passwords)", ErrPrecondition, cur.Username)
}
