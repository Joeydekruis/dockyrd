package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type variableResponse struct {
	Key       string  `json:"key"`
	IsSecret  bool    `json:"is_secret"`
	AtBuild   bool    `json:"at_build"`
	AtRuntime bool    `json:"at_runtime"`
	Value     string  `json:"value,omitempty"`
	LinkID    *string `json:"link_id,omitempty"`
}

type putVariablesRequest struct {
	Variables []struct {
		Key       string `json:"key"`
		Value     string `json:"value"`
		IsSecret  bool   `json:"is_secret"`
		AtBuild   bool   `json:"at_build"`
		AtRuntime bool   `json:"at_runtime"`
	} `json:"variables"`
}

func registerVariableRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/variables", deps.handleListVariables)
	mux.HandleFunc("PUT /api/v1/services/{id}/variables", deps.handlePutVariables)
}

func (d apiDeps) handleListVariables(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetServiceByID(ctx, serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	rows, err := d.q.ListVariablesForService(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	reveal := r.URL.Query().Get("reveal") == "1"
	out := make([]variableResponse, 0, len(rows))
	for _, row := range rows {
		item := variableResponse{
			Key:       row.Key,
			IsSecret:  row.IsSecret != 0,
			AtBuild:   row.AtBuild != 0,
			AtRuntime: row.AtRuntime != 0,
		}
		if row.LinkID.Valid {
			id := row.LinkID.String
			item.LinkID = &id
		}
		if reveal || row.IsSecret == 0 {
			plain, err := crypto.Decrypt(key, row.ValueEnc)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "decrypt error")
				return
			}
			item.Value = string(plain)
		} else if row.IsSecret != 0 {
			item.Value = "••••••••"
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"variables": out})
}

func (d apiDeps) handlePutVariables(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetServiceByID(ctx, serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	var body putVariablesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	var vals []string
	for _, v := range body.Variables {
		k := strings.TrimSpace(v.Key)
		if k == "" {
			continue
		}
		vals = append(vals, v.Value)
	}
	if err := validateVariableRefs(ctx, d.q, serviceID, vals); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := d.q.DeleteVariablesForService(ctx, serviceID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	for _, v := range body.Variables {
		k := strings.TrimSpace(v.Key)
		if k == "" {
			continue
		}
		enc, err := crypto.Encrypt(masterKey, []byte(v.Value))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "encrypt error")
			return
		}
		atBuild := int64(1)
		if !v.AtBuild {
			atBuild = 0
		}
		atRuntime := int64(1)
		if !v.AtRuntime {
			atRuntime = 0
		}
		isSecret := int64(0)
		if v.IsSecret {
			isSecret = 1
		}
		if err := d.q.InsertVariable(ctx, sqlc.InsertVariableParams{
			ID:        ulid.Make().String(),
			ServiceID: serviceID,
			Key:       k,
			ValueEnc:  enc,
			IsSecret:  isSecret,
			AtBuild:   atBuild,
			AtRuntime: atRuntime,
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "insert variable error")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
