package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

func registerGitHubRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/github/status", deps.handleGitHubStatus)
	mux.HandleFunc("DELETE /api/v1/github/connection", deps.handleGitHubDisconnect)
	mux.HandleFunc("POST /api/v1/github/app/manifest", deps.handleGitHubManifest)
	mux.HandleFunc("GET /api/v1/github/app/callback", deps.handleGitHubCallback)
	mux.HandleFunc("GET /api/v1/github/repos", deps.handleGitHubRepos)
	mux.HandleFunc("GET /api/v1/github/repos/{owner}/{repo}/branches", deps.handleGitHubBranches)
	mux.HandleFunc("POST /api/v1/hooks/github", deps.handleGitHubWebhook)
}

func (d apiDeps) handleGitHubStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	st, err := github.LoadStatus(r.Context(), d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	base := d.githubPublicBaseURL(r)
	st, _ = github.EnrichStatus(r.Context(), d.q, d.cfg.SecretKeyPath(), base, st)
	writeJSON(w, http.StatusOK, st)
}

func (d apiDeps) handleGitHubDisconnect(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	if err := d.q.DeleteGitSource(r.Context(), github.DefaultGitSourceID()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "disconnect failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleGitHubManifest(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	var body struct {
		AppName      string `json:"app_name"`
		PublicURL    string `json:"public_url"`
		Organization string `json:"organization"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	base := strings.TrimRight(strings.TrimSpace(body.PublicURL), "/")
	if base == "" {
		base = d.githubPublicBaseURL(r)
	}
	state := github.NewStateToken()
	if d.q != nil {
		val, _ := json.Marshal(map[string]string{"state": state})
		_ = d.q.UpsertSetting(r.Context(), sqlc.UpsertSettingParams{
			Key:   "github.manifest_state",
			Value: val,
		})
	}
	launch, err := github.BuildManifestLaunch(github.ManifestParams{
		PublicBaseURL: base,
		AppName:       body.AppName,
		Organization:  body.Organization,
	}, state)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, launch)
}

func (d apiDeps) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "code required")
		return
	}
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	if state := strings.TrimSpace(r.URL.Query().Get("state")); state != "" {
		setting, err := d.q.GetSetting(r.Context(), "github.manifest_state")
		if err == nil {
			var stored struct {
				State string `json:"state"`
			}
			raw, _ := json.Marshal(setting.Value)
			_ = json.Unmarshal(raw, &stored)
			if stored.State != "" && stored.State != state {
				writeJSONError(w, http.StatusBadRequest, "invalid oauth state")
				return
			}
		}
	}
	res, err := github.ConvertManifest(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github conversion failed")
		return
	}
	installID := int64(0)
	if v := strings.TrimSpace(r.URL.Query().Get("installation_id")); v != "" {
		installID, _ = strconv.ParseInt(v, 10, 64)
	}
	if err := github.SaveConversion(r.Context(), d.q, d.cfg.SecretKeyPath(), res, installID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "store github app failed")
		return
	}
	if installID > 0 {
		_ = github.SaveInstallationID(r.Context(), d.q, installID)
	}
	http.Redirect(w, r, "/settings/server/integrations?github=app-created", http.StatusFound)
}

func (d apiDeps) handleGitHubRepos(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	token, err := d.githubInstallationToken(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	repos, err := github.ListRepos(r.Context(), token, r.URL.Query().Get("q"))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github api error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repos": repos})
}

func (d apiDeps) handleGitHubBranches(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	token, err := d.githubInstallationToken(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")
	branches, err := github.ListBranches(r.Context(), token, owner, repo)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github api error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (d apiDeps) githubInstallationToken(r *http.Request) (string, error) {
	creds, err := github.LoadCredentials(r.Context(), d.q, d.cfg.SecretKeyPath())
	if err != nil {
		return "", err
	}
	return creds.InstallationAccessToken(r.Context())
}

func (d apiDeps) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	secret, err := github.WebhookSecretPlain(r.Context(), d.q, d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "webhook config error")
		return
	}
	if !github.VerifyWebhookSignature(secret, body, r.Header.Get("X-Hub-Signature-256")) {
		writeJSONError(w, http.StatusUnauthorized, "invalid signature")
		return
	}
	eventType := r.Header.Get("X-GitHub-Event")
	switch eventType {
	case "installation":
		ev, err := github.ParseInstallationEvent(body)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid installation payload")
			return
		}
		switch strings.ToLower(ev.Action) {
		case "created", "added", "new_permissions_accepted":
			if err := github.SaveInstallationID(r.Context(), d.q, ev.Installation.ID); err != nil {
				writeJSONError(w, http.StatusInternalServerError, "store installation failed")
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
		return
	case "push":
		// handled below
	default:
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ev, err := github.ParsePushEvent(body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid push payload")
		return
	}
	branch := github.BranchFromRef(ev.Ref)
	serviceIDs, err := github.ListMatchingServices(r.Context(), d.q, ev.Repository.FullName, branch)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "match services error")
		return
	}
	meta := jobs.GitMeta{
		CommitSHA:     ev.After,
		CommitMessage: ev.HeadCommit.Message,
		CommitAuthor:  ev.HeadCommit.Author.Name,
		Branch:        branch,
	}
	for _, id := range serviceIDs {
		deploymentID, err := jobs.EnqueueServiceDeploy(r.Context(), d.q, id, "github_push", "", meta)
		if err == nil && d.events != nil {
			d.events.Publish("deployments:"+deploymentID, map[string]string{"status": "queued"})
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func publicBaseURL(r *http.Request) string {
	if v := strings.TrimRight(os.Getenv("DOCKYARD_PUBLIC_URL"), "/"); v != "" {
		return v
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = strings.TrimSpace(fwd)
	}
	return scheme + "://" + r.Host
}

func (d apiDeps) githubPublicBaseURL(r *http.Request) string {
	if d.q != nil {
		panel, err := instance.LoadPanel(r.Context(), d.q)
		if err == nil {
			if v := instance.PublicBaseURL(d.cfg, panel); v != "" {
				return v
			}
		}
	}
	return publicBaseURL(r)
}
