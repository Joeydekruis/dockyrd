package addons

import (
	"errors"

	"github.com/Joeydekruis/dockyrd/internal/hostops"
)

// Kind groups catalog entries.
type Kind string

const (
	KindBaseline    Kind = "baseline"
	KindRecommended Kind = "recommended"
	KindOptional    Kind = "optional"
)

// CatalogEntry is a static add-on definition.
type CatalogEntry struct {
	ID          string
	Title       string
	Description string
	Kind        Kind
	Installable bool
	HostAction  string
	Version     string
}

var catalog = []CatalogEntry{
	{
		ID: "firewall", Title: "Host firewall", Description: "Baseline nftables rules (22, 80, 443)",
		Kind: KindBaseline, Installable: true, HostAction: hostops.KindFirewallEnsure, Version: "1",
	},
	{
		ID: "unattended-updates", Title: "Unattended updates", Description: "Security updates without auto-reboot",
		Kind: KindBaseline, Installable: true, HostAction: hostops.KindUnattendedEnable, Version: "1",
	},
	{
		ID: "crowdsec", Title: "CrowdSec", Description: "Intrusion prevention (coming in P3)",
		Kind: KindRecommended, Installable: false,
	},
	{
		ID: "proxy-waf", Title: "Proxy WAF", Description: "Coraza WAF via proxybuild (P2)",
		Kind: KindOptional, Installable: false,
	},
	{
		ID: "proxy-rate-limit", Title: "Proxy rate limiting", Description: "Edge rate limits via proxybuild (P2)",
		Kind: KindOptional, Installable: false,
	},
}

// Catalog returns all known add-ons.
func Catalog() []CatalogEntry {
	out := make([]CatalogEntry, len(catalog))
	copy(out, catalog)
	return out
}

// ByID returns a catalog entry or error.
func ByID(id string) (CatalogEntry, error) {
	for _, e := range catalog {
		if e.ID == id {
			return e, nil
		}
	}
	return CatalogEntry{}, errors.New("unknown addon")
}
