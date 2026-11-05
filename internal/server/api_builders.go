package server

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
)

func (d apiDeps) handleListBuilders(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rows, err := d.q.ListBuilders(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id":           row.ID,
			"name":         row.Name,
			"kind":         row.Kind,
			"endpoint":     row.Endpoint,
			"platforms":    nullStrPtr(row.Platforms),
			"status":       nullStrPtr(row.Status),
			"last_seen_at": nullInt64Ptr(row.LastSeenAt),
			"created_at":   row.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"builders": out})
}

func (d apiDeps) handleCreateBuilder(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		Name     string `json:"name"`
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(body.Name)
	endpoint := strings.TrimSpace(body.Endpoint)
	if name == "" || endpoint == "" {
		writeJSONError(w, http.StatusBadRequest, "name and endpoint required")
		return
	}
	if !strings.HasPrefix(endpoint, "tcp://") {
		endpoint = "tcp://" + endpoint
	}
	host := hostFromEndpoint(endpoint)
	bundle, err := build.GenerateRemoteCertBundle(host)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "certificate generation failed")
		return
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "master key unavailable")
		return
	}
	caEnc, _ := crypto.Encrypt(key, bundle.CA)
	clientCertEnc, _ := crypto.Encrypt(key, bundle.ClientCert)
	clientKeyEnc, _ := crypto.Encrypt(key, bundle.ClientKey)
	id := ulid.Make().String()
	now := time.Now().Unix()
	if err := d.q.InsertBuilder(r.Context(), sqlc.InsertBuilderParams{
		ID:            id,
		Name:          name,
		Kind:          "remote_buildkit",
		Endpoint:      endpoint,
		CaCertEnc:     caEnc,
		ClientCertEnc: clientCertEnc,
		ClientKeyEnc:  clientKeyEnc,
		Platforms:     sql.NullString{},
		Status:        sql.NullString{String: "pending", Valid: true},
		LastSeenAt:    sql.NullInt64{},
		CreatedAt:     now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "store builder failed")
		return
	}
	certDir := filepath.Join(d.cfg.DataDir, "builders", id)
	_ = build.WriteCertDir(certDir, bundle)
	_ = os.WriteFile(filepath.Join(certDir, "server-cert.pem"), bundle.ServerCert, 0o644)
	_ = os.WriteFile(filepath.Join(certDir, "server-key.pem"), bundle.ServerKey, 0o600)
	port := portFromEndpoint(endpoint)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":       id,
		"name":     name,
		"endpoint": endpoint,
		"docker_run": build.RemoteBuilderRunLine(port),
		"server_certs": map[string]string{
			"ca.pem":   string(bundle.CA),
			"cert.pem": string(bundle.ServerCert),
			"key.pem":  string(bundle.ServerKey),
		},
	})
}

func (d apiDeps) handleDeleteBuilder(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if err := d.q.DeleteBuilder(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	_ = os.RemoveAll(filepath.Join(d.cfg.DataDir, "builders", id))
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleTestBuilder(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	row, err := d.q.GetBuilder(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "builder not found")
		return
	}
	certDir := filepath.Join(d.cfg.DataDir, "builders", id)
	builderName := "dy-remote-" + id
	if err := build.EnsureRemoteBuildx(r.Context(), builderName, row.Endpoint, certDir); err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	out, err := build.TestRemoteBuilder(r.Context(), builderName)
	now := time.Now().Unix()
	status := "online"
	if err != nil {
		status = "error"
		_ = d.q.UpdateBuilderStatus(r.Context(), sqlc.UpdateBuilderStatusParams{
			Status:     sql.NullString{String: status, Valid: true},
			Platforms:  sql.NullString{String: err.Error(), Valid: true},
			LastSeenAt: sql.NullInt64{Int64: now, Valid: true},
			ID:         id,
		})
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	_ = d.q.UpdateBuilderStatus(r.Context(), sqlc.UpdateBuilderStatusParams{
		Status:     sql.NullString{String: status, Valid: true},
		Platforms:  sql.NullString{String: "linux/amd64", Valid: true},
		LastSeenAt: sql.NullInt64{Int64: now, Valid: true},
		ID:         id,
	})
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "detail": out})
}

func (d apiDeps) handleBuilderBundle(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	row, err := d.q.GetBuilder(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "builder not found")
		return
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "master key unavailable")
		return
	}
	ca, _ := crypto.Decrypt(key, row.CaCertEnc)
	clientCert, _ := crypto.Decrypt(key, row.ClientCertEnc)
	clientKey, _ := crypto.Decrypt(key, row.ClientKeyEnc)
	certDir := filepath.Join(d.cfg.DataDir, "builders", id)
	serverCert, _ := os.ReadFile(filepath.Join(certDir, "server-cert.pem"))
	serverKey, _ := os.ReadFile(filepath.Join(certDir, "server-key.pem"))
	writeJSON(w, http.StatusOK, map[string]any{
		"endpoint": row.Endpoint,
		"files": map[string]string{
			"ca.pem":   string(ca),
			"cert.pem": string(serverCert),
			"key.pem":  string(serverKey),
		},
		"client_cert.pem": string(clientCert),
		"client_key.pem":  string(clientKey),
		"docker_run":      build.RemoteBuilderRunLine(portFromEndpoint(row.Endpoint)),
	})
}

func (d apiDeps) handlePatchBuildLocation(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	var body struct {
		BuildLocation string `json:"build_location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	loc := strings.TrimSpace(body.BuildLocation)
	if loc == "" {
		writeJSONError(w, http.StatusBadRequest, "build_location required")
		return
	}
	raw, err := github.LoadServiceConfig(r.Context(), d.q, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "service not found")
		return
	}
	svc, err := d.q.GetServiceByID(r.Context(), serviceID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "service not found")
		return
	}
	var merged []byte
	switch svc.Kind {
	case "app":
		cfg, err := appconfig.Parse(raw)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid app config")
			return
		}
		cfg.BuildLocation = loc
		merged, err = json.Marshal(cfg)
	case "static_site":
		var cfg map[string]any
		_ = json.Unmarshal(raw, &cfg)
		if cfg == nil {
			cfg = map[string]any{}
		}
		cfg["buildLocation"] = loc
		merged, err = json.Marshal(cfg)
	default:
		writeJSONError(w, http.StatusBadRequest, "build location applies to app and static_site")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encode error")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpdateServiceConfig(r.Context(), sqlc.UpdateServiceConfigParams{
		Config:    merged,
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        serviceID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"build_location": loc})
}

func hostFromEndpoint(endpoint string) string {
	endpoint = strings.TrimPrefix(strings.TrimSpace(endpoint), "tcp://")
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return endpoint
	}
	return host
}

func portFromEndpoint(endpoint string) string {
	endpoint = strings.TrimPrefix(strings.TrimSpace(endpoint), "tcp://")
	_, port, err := net.SplitHostPort(endpoint)
	if err != nil || port == "" {
		return "1234:1234"
	}
	return port + ":" + port
}
