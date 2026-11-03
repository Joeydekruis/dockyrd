package server

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
)

func registerBucketKeyRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/keys", deps.handleListBucketKeys)
	mux.HandleFunc("POST /api/v1/services/{id}/keys", deps.handleCreateBucketKey)
	mux.HandleFunc("DELETE /api/v1/services/{id}/keys/{keyId}", deps.handleDeleteBucketKey)
}

type bucketKeyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	AccessKeyID string `json:"access_key_id"`
	Secret      string `json:"secret,omitempty"`
}

func (d apiDeps) handleListBucketKeys(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	_, cfg, masterKey, ok := d.loadBucketService(w, r)
	if !ok {
		return
	}
	reveal := r.URL.Query().Get("reveal") == "1"
	out := make([]bucketKeyResponse, 0, len(cfg.Keys))
	for _, k := range cfg.Keys {
		item := bucketKeyResponse{ID: k.ID, Name: k.Name, AccessKeyID: k.AccessKeyID}
		if reveal {
			sec, err := dbkind.DecryptKeySecret(masterKey, k)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "decrypt error")
				return
			}
			item.Secret = sec
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": out})
}

func (d apiDeps) handleCreateBucketKey(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	svc, cfg, masterKey, ok := d.loadBucketService(w, r)
	if !ok {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	accessKeyID, secret, err := dbkind.NewAccessKeyPair()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "key generation failed")
		return
	}
	secretEnc, err := crypto.Encrypt(masterKey, []byte(secret))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encrypt error")
		return
	}
	newKey := dbkind.BucketKey{
		ID:          ulid.Make().String(),
		Name:        strings.TrimSpace(body.Name),
		AccessKeyID: accessKeyID,
		SecretEnc:   base64.StdEncoding.EncodeToString(secretEnc),
	}
	cfg.Keys = append(cfg.Keys, newKey)
	if err := d.saveBucketConfig(r, svc.ID, cfg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save error")
		return
	}
	writeJSON(w, http.StatusCreated, bucketKeyResponse{
		ID:          newKey.ID,
		Name:        newKey.Name,
		AccessKeyID: newKey.AccessKeyID,
		Secret:      secret,
	})
}

func (d apiDeps) handleDeleteBucketKey(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	svc, cfg, _, ok := d.loadBucketService(w, r)
	if !ok {
		return
	}
	keyID := r.PathValue("keyId")
	next := make([]dbkind.BucketKey, 0, len(cfg.Keys))
	found := false
	for _, k := range cfg.Keys {
		if k.ID == keyID {
			found = true
			continue
		}
		next = append(next, k)
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "key not found")
		return
	}
	if len(next) == 0 {
		writeJSONError(w, http.StatusBadRequest, "cannot remove last key")
		return
	}
	cfg.Keys = next
	if err := d.saveBucketConfig(r, svc.ID, cfg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) loadBucketService(w http.ResponseWriter, r *http.Request) (sqlc.Service, dbkind.BucketConfig, []byte, bool) {
	serviceID := r.PathValue("id")
	row, err := d.q.GetServiceByID(r.Context(), serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return sqlc.Service{}, dbkind.BucketConfig{}, nil, false
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return sqlc.Service{}, dbkind.BucketConfig{}, nil, false
	}
	if row.Kind != "bucket" {
		writeJSONError(w, http.StatusBadRequest, "not a bucket service")
		return sqlc.Service{}, dbkind.BucketConfig{}, nil, false
	}
	masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return sqlc.Service{}, dbkind.BucketConfig{}, nil, false
	}
	cfg, err := dbkind.ParseBucket(serviceConfigJSON(row.Config))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "config error")
		return sqlc.Service{}, dbkind.BucketConfig{}, nil, false
	}
	return row, cfg, masterKey, true
}

func (d apiDeps) saveBucketConfig(r *http.Request, serviceID string, cfg dbkind.BucketConfig) error {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return d.q.UpdateServiceConfig(r.Context(), sqlc.UpdateServiceConfigParams{
		Config:    raw,
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        serviceID,
	})
}
