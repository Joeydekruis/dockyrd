package doctor

import (
	"encoding/json"
	"os"
)

const dockerDaemonPath = "/etc/docker/daemon.json"

// checkDockerAddressPools warns when default-address-pools is missing (needed for many dy-l-* link networks).
func checkDockerAddressPools() []Check {
	raw, err := os.ReadFile(dockerDaemonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Check{{
				Name:    "Docker address pools",
				Status:  StatusWarn,
				Detail:  "no /etc/docker/daemon.json",
				FixHint: `Merge "default-address-pools":[{"base":"10.200.0.0/16","size":27}] into daemon.json before heavy link usage`,
			}}
		}
		return nil
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return []Check{{
			Name:   "Docker address pools",
			Status: StatusWarn,
			Detail: "could not parse daemon.json",
		}}
	}
	if _, ok := doc["default-address-pools"]; ok {
		return []Check{{
			Name:   "Docker address pools",
			Status: StatusOK,
			Detail: "default-address-pools configured",
		}}
	}
	return []Check{{
		Name:    "Docker address pools",
		Status:  StatusWarn,
		Detail:  "default-address-pools not set",
		FixHint: `Add "default-address-pools":[{"base":"10.200.0.0/16","size":27}] to /etc/docker/daemon.json and restart docker when safe`,
	}}
}
