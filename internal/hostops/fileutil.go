package hostops

import "os"

func restoreHostFile(env HostEnv, path, previous string) error {
	if previous == "" {
		return os.Remove(path)
	}
	return env.WriteFile(path, []byte(previous), 0o644)
}
