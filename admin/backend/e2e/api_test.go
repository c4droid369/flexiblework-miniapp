// Package e2e drives the running backend over real HTTP. It assumes
// `docker compose up -d` is already serving the API on $E2E_BASE_URL
// (default http://localhost:8080).
//
// Run with:  go test -tags=e2e ./e2e/...
//
// The tests intentionally do NOT mock anything — they exercise the same
// code paths as a browser, including JWT validation, RBAC, refresh-token
// rotation, and the operation-log middleware.
//
//nolint:errcheck // e2e tests intentionally close bodies and discard helpers
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultBaseURL  = "http://localhost:8080"
	adminUser       = "admin"
	adminPassword   = "admin123"
	probeTimeout    = 60 * time.Second
	shortRetryEvery = 500 * time.Millisecond
)

func baseURL() string {
	if v := os.Getenv("E2E_BASE_URL"); v != "" {
		return v
	}
	return defaultBaseURL
}

// waitForBackend polls /healthz until it returns 200 or the deadline
// elapses. Skips the test (t.Skip) if the stack isn't up, instead of
// failing — so CI without docker compose doesn't break.
func waitForBackend(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(probeTimeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL() + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(shortRetryEvery)
	}
	t.Skipf("backend not reachable at %s after %s — run `docker compose up -d`", baseURL(), probeTimeout)
}

type apiClient struct {
	t       *testing.T
	baseURL string
	token   string
	refresh string
}

func newClient(t *testing.T) *apiClient {
	t.Helper()
	waitForBackend(t)
	c := &apiClient{t: t, baseURL: baseURL()}
	c.login(adminUser, adminPassword)
	return c
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *apiClient) do(method, path string, body any, withAuth bool) (int, envelope) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		c.t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if withAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	var env envelope
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &env)
	}
	return resp.StatusCode, env
}

func (c *apiClient) login(user, pass string) {
	c.t.Helper()
	status, env := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": user, "password": pass,
	}, false)
	if status != 200 || env.Code != 0 {
		c.t.Fatalf("login failed: status=%d body=%+v", status, env)
	}
	var d struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		Permissions  []string `json:"permissions"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil {
		c.t.Fatalf("decode login response: %v", err)
	}
	c.token = d.AccessToken
	c.refresh = d.RefreshToken
	if len(d.Permissions) == 0 {
		c.t.Fatalf("login returned no permissions — RBAC broken?")
	}
}

func TestE2E_Health(t *testing.T) {
	waitForBackend(t)
	resp, err := http.Get(baseURL() + "/healthz")
	if err != nil {
		t.Fatalf("healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("healthz status: %d", resp.StatusCode)
	}
}

func TestE2E_LoginAndMe(t *testing.T) {
	c := newClient(t)

	// /auth/me should return the admin profile with at least one role and
	// some permission codes.
	status, env := c.do(http.MethodGet, "/api/v1/auth/me", nil, true)
	if status != 200 || env.Code != 0 {
		t.Fatalf("me: status=%d body=%+v", status, env)
	}
	var me map[string]any
	_ = json.Unmarshal(env.Data, &me)
	if me["username"] != "admin" {
		t.Fatalf("me username: got %v", me["username"])
	}
	roles, _ := me["roles"].([]any)
	if len(roles) == 0 {
		t.Fatalf("me has no roles: %+v", me)
	}
}

func TestE2E_LoginInvalidPassword_401(t *testing.T) {
	waitForBackend(t)
	status, env := (&apiClient{t: t, baseURL: baseURL()}).do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": adminUser, "password": "wrong-password",
	}, false)
	if status != 401 {
		t.Fatalf("expected 401 for wrong password, got %d", status)
	}
	if env.Code == 0 {
		t.Fatalf("expected non-zero code in envelope: %+v", env)
	}
}

func TestE2E_RefreshToken_Rotation(t *testing.T) {
	c := newClient(t)
	oldToken := c.token
	oldRefresh := c.refresh

	status, env := c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": oldRefresh,
	}, false)
	if status != 200 || env.Code != 0 {
		t.Fatalf("refresh: status=%d body=%+v", status, env)
	}
	var d struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(env.Data, &d)
	if d.AccessToken == oldToken {
		t.Fatalf("refresh returned the same access token (rotation broken)")
	}
	if d.RefreshToken == oldRefresh {
		t.Fatalf("refresh returned the same refresh token (rotation broken)")
	}
	c.token = d.AccessToken
	c.refresh = d.RefreshToken
}

func TestE2E_RefreshReuse_RevokesFamily(t *testing.T) {
	c := newClient(t)
	oldRefresh := c.refresh

	// Rotate once.
	_, env := c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": oldRefresh,
	}, false)
	if env.Code != 0 {
		t.Fatalf("first refresh: %+v", env)
	}

	// Reuse the OLD refresh — must be rejected, and the family must be
	// revoked so the latest refresh no longer works.
	status, env := c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": oldRefresh,
	}, false)
	if status != 401 {
		t.Fatalf("reuse: expected 401, got %d body=%+v", status, env)
	}

	// The current refresh must now also be invalid.
	status, env = c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": c.refresh,
	}, false)
	if status != 401 {
		t.Fatalf("family revocation: expected 401, got %d body=%+v", status, env)
	}
}

func TestE2E_UserCRUDLifecycle(t *testing.T) {
	c := newClient(t)

	// List users (paginated).
	status, env := c.do(http.MethodGet, "/api/v1/system/users?page=1&size=10", nil, true)
	if status != 200 || env.Code != 0 {
		t.Fatalf("list users: status=%d body=%+v", status, env)
	}

	// Create a user with a unique username so re-runs are clean.
	uname := fmt.Sprintf("e2e_%d", time.Now().UnixNano())
	status, env = c.do(http.MethodPost, "/api/v1/system/users", map[string]any{
		"username": uname,
		"password": "test-pass-123",
		"nickname": "E2E user",
		"email":    "e2e@example.com",
	}, true)
	if status != 201 || env.Code != 0 {
		t.Fatalf("create user: status=%d body=%+v", status, env)
	}
	var created struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(env.Data, &created)
	if created.ID == 0 {
		t.Fatalf("create returned id=0")
	}

	// Read back.
	status, env = c.do(http.MethodGet, fmt.Sprintf("/api/v1/system/users/%d", created.ID), nil, true)
	if status != 200 {
		t.Fatalf("get user: status=%d", status)
	}

	// Reset password.
	status, env = c.do(http.MethodPost, fmt.Sprintf("/api/v1/system/users/%d/reset-password", created.ID), map[string]string{
		"new_password": "new-pass-456",
	}, true)
	if status != 200 {
		t.Fatalf("reset password: status=%d body=%+v", status, env)
	}

	// Delete.
	status, env = c.do(http.MethodDelete, fmt.Sprintf("/api/v1/system/users/%d", created.ID), nil, true)
	if status != 200 {
		t.Fatalf("delete user: status=%d body=%+v", status, env)
	}
}

func TestE2E_DuplicateUsername_409(t *testing.T) {
	c := newClient(t)
	body := map[string]any{
		"username": "admin", // already exists from seed
		"password": "whatever",
	}
	status, env := c.do(http.MethodPost, "/api/v1/system/users", body, true)
	if status != 409 {
		t.Fatalf("expected 409 for duplicate username, got %d body=%+v", status, env)
	}
}

func TestE2E_RoleAssignMenu(t *testing.T) {
	c := newClient(t)

	// Create a role.
	roleCode := fmt.Sprintf("e2e_role_%d", time.Now().UnixNano())
	status, env := c.do(http.MethodPost, "/api/v1/system/roles", map[string]any{
		"code": roleCode,
		"name": "E2E Role",
	}, true)
	if status != 201 {
		t.Fatalf("create role: status=%d body=%+v", status, env)
	}
	var created struct {
		ID uint64 `json:"id"`
	}
	_ = json.Unmarshal(env.Data, &created)

	// Get the menu tree and pick a menu id.
	_, env = c.do(http.MethodGet, "/api/v1/system/menus", nil, true)
	if env.Code != 0 {
		t.Fatalf("list menus: %+v", env)
	}
	var menus []struct {
		ID       uint64 `json:"id"`
		Type     int    `json:"type"`
		PermCode string `json:"perm_code"`
	}
	_ = json.Unmarshal(env.Data, &menus)
	if len(menus) == 0 {
		t.Fatalf("no menus returned")
	}
	var menuIDs []uint64
	for _, m := range menus {
		if m.Type == 2 || m.Type == 3 { // Menu or Button
			menuIDs = append(menuIDs, m.ID)
		}
	}
	if len(menuIDs) == 0 {
		t.Fatalf("no routable menus")
	}

	// Assign menus.
	status, env = c.do(http.MethodPost, fmt.Sprintf("/api/v1/system/roles/%d/menus", created.ID), map[string]any{
		"menu_ids": menuIDs,
	}, true)
	if status != 200 {
		t.Fatalf("assign menus: status=%d body=%+v", status, env)
	}

	// Verify assignment via GET role.
	status, env = c.do(http.MethodGet, fmt.Sprintf("/api/v1/system/roles/%d", created.ID), nil, true)
	if status != 200 {
		t.Fatalf("get role: status=%d", status)
	}

	// Cleanup: delete role.
	status, _ = c.do(http.MethodDelete, fmt.Sprintf("/api/v1/system/roles/%d", created.ID), nil, true)
	if status != 200 {
		t.Logf("cleanup: delete role status=%d (best-effort)", status)
	}
}

func TestE2E_LogoutRevokesRefresh(t *testing.T) {
	c := newClient(t)

	// Logout (this revokes the entire refresh-token family).
	status, _ := c.do(http.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": c.refresh,
	}, true)
	if status != 200 {
		t.Fatalf("logout status=%d", status)
	}

	// Try to use the old refresh — should fail.
	status, _ = c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": c.refresh,
	}, false)
	if status != 401 {
		t.Fatalf("after logout, refresh should be 401, got %d", status)
	}
}

func TestE2E_FileUploadAndDownload(t *testing.T) {
	c := newClient(t)

	// Build a multipart body.
	var buf bytes.Buffer
	w := multipartWriter(&buf, "file", "hello.txt", "hello e2e")
	w.Close()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/upload", &buf)
	if err != nil {
		t.Fatalf("build upload: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload status=%d body=%s", resp.StatusCode, body)
	}
	var env envelope
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &env)
	if env.Code != 0 {
		t.Fatalf("upload envelope: %+v", env)
	}
	var d struct {
		ID  uint64 `json:"id"`
		URL string `json:"url"`
	}
	_ = json.Unmarshal(env.Data, &d)
	if d.ID == 0 || d.URL == "" {
		t.Fatalf("upload returned no id/url: %+v", d)
	}

	// Fetch the file back via the returned URL (already absolute).
	fileResp, err := http.Get(d.URL)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer fileResp.Body.Close()
	if fileResp.StatusCode != 200 {
		t.Fatalf("download status=%d", fileResp.StatusCode)
	}
	got, _ := io.ReadAll(fileResp.Body)
	if !bytes.Equal(got, []byte("hello e2e")) {
		t.Fatalf("download body mismatch: got %q", got)
	}

	// Cleanup.
	_, _ = c.do(http.MethodDelete, fmt.Sprintf("/api/v1/files-list/%d", d.ID), nil, true)
}

func TestE2E_ExportExcel(t *testing.T) {
	c := newClient(t)

	req, _ := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/system/users/export/excel", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("export status=%d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" || !contains(ct, "spreadsheetml") {
		t.Fatalf("unexpected content-type: %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) < 100 {
		t.Fatalf("xlsx body suspiciously small: %d bytes", len(body))
	}
}

func TestE2E_OperationLog_RecordsRequests(t *testing.T) {
	c := newClient(t)

	// Hit a few endpoints; the operation log middleware should record them.
	c.do(http.MethodGet, "/api/v1/system/users?page=1&size=5", nil, true)
	c.do(http.MethodGet, "/api/v1/system/roles?page=1&size=5", nil, true)

	// Query the log; we expect at least our two requests to appear.
	status, env := c.do(http.MethodGet, "/api/v1/system/logs?page=1&size=20&keyword=system/users", nil, true)
	if status != 200 || env.Code != 0 {
		t.Fatalf("log query: status=%d body=%+v", status, env)
	}
	var page struct {
		List  []map[string]any `json:"list"`
		Total int64            `json:"total"`
	}
	_ = json.Unmarshal(env.Data, &page)
	if len(page.List) == 0 {
		t.Fatalf("operation log empty — middleware broken?")
	}
}

// contains is a tiny helper to avoid pulling in strings for a single check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
