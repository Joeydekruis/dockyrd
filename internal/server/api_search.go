package server

import (
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerSearchRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/search", deps.handleSearch)
}

type searchHit struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Href  string `json:"href"`
}

var searchNavPages = []searchHit{
	{Kind: "page", Label: "Dashboard", Href: "/"},
	{Kind: "page", Label: "Projects", Href: "/projects"},
	{Kind: "page", Label: "Applications", Href: "/applications"},
	{Kind: "page", Label: "Databases", Href: "/databases"},
	{Kind: "page", Label: "Static sites", Href: "/static-sites"},
	{Kind: "page", Label: "Object storage", Href: "/object-storage"},
	{Kind: "page", Label: "Analytics", Href: "/analytics"},
	{Kind: "page", Label: "Profile & security", Href: "/settings/profile"},
	{Kind: "page", Label: "Team", Href: "/settings/team"},
	{Kind: "page", Label: "Updates", Href: "/settings/updates"},
	{Kind: "page", Label: "Instance settings", Href: "/settings/instance"},
}

func (d apiDeps) handleSearch(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": searchNavPages})
		return
	}
	needle := strings.ToLower(q)
	out := make([]searchHit, 0, 16)
	for _, page := range searchNavPages {
		if strings.Contains(strings.ToLower(page.Label), needle) {
			out = append(out, page)
		}
	}
	pattern := likePattern(q)
	rows, err := d.q.SearchProjects(r.Context(), sqlc.SearchProjectsParams{
		Name: pattern,
		Slug: pattern,
	})
	if err == nil {
		for _, row := range rows {
			out = append(out, searchHit{
				Kind:  "project",
				Label: row.Name,
				Href:  "/projects/" + row.Slug,
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}

func likePattern(q string) string {
	q = strings.ReplaceAll(q, "%", "")
	q = strings.ReplaceAll(q, "_", "")
	return "%" + q + "%"
}
