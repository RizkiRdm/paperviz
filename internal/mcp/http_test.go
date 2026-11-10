package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// newHTTPTestDB opens a test DB with two users holding distinct service
// keys. Keys return in plaintext only to the test that created them.
func newHTTPTestDB(t *testing.T) (*sql.DB, string, string) {
	t.Helper()
	db := newTestDB(t)
	for _, u := range []struct{ id, key string }{{"user-a", "pv-test-key-a"}, {"user-b", "pv-test-key-b"}} {
		_, err := db.Exec(
			`INSERT INTO users (id, email, password_hash, created_at, api_key_hash) VALUES (?, ?, ?, ?, ?)`,
			u.id, u.id+"@example.test", "not-used", 0, hashPresentedKey(u.key),
		)
		if err != nil {
			t.Fatalf("seed user %s: %v", u.id, err)
		}
	}
	return db, "pv-test-key-a", "pv-test-key-b"
}

// newHTTPServer stands up the remote transport on a test server.
func newHTTPServer(t *testing.T) (*httptest.Server, string, string) {
	t.Helper()
	db, keyA, keyB := newHTTPTestDB(t)
	ts := httptest.NewServer(NewHTTPServer(db).Handler())
	t.Cleanup(ts.Close)
	return ts, keyA, keyB
}

// bearerTransport injects the test key as a Bearer credential on every call.
type bearerTransport struct {
	key   string
	inner http.RoundTripper
}

// RoundTrip clones the request and stamps the Authorization header.
func (t *bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+t.key)
	return t.inner.RoundTrip(r)
}

// connectHTTPClient completes a real initialize handshake over HTTP.
func connectHTTPClient(t *testing.T, endpoint, key string) (*sdk.ClientSession, context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	client := sdk.NewClient(&sdk.Implementation{Name: "http-test", Version: "0.0.1"}, nil)
	transport := &sdk.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: &http.Client{Transport: &bearerTransport{key: key, inner: http.DefaultTransport}},
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		cancel()
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session, ctx, cancel
}

// callToolText runs one tool call and returns its first text content. A
// tool-level error (result.IsError) is returned as text, not a Go error,
// so tests can assert on it uniformly.
func callToolText(t *testing.T, session *sdk.ClientSession, ctx context.Context, name string, args map[string]any) string {
	t.Helper()
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	content, ok := result.Content[0].(*sdk.TextContent)
	if !ok {
		t.Fatalf("%s: content type = %T, want *sdk.TextContent", name, result.Content[0])
	}
	return content.Text
}

// ingestText ingests a paper body under the client's identity and returns
// the document_id.
func ingestText(t *testing.T, session *sdk.ClientSession, ctx context.Context, text string) string {
	t.Helper()
	out := callToolText(t, session, ctx, "ingest_document", map[string]any{"text": text})
	var decoded struct {
		DocumentID string `json:"document_id"`
	}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil || decoded.DocumentID == "" {
		t.Fatalf("no document_id in %s (err=%v)", out, err)
	}
	return decoded.DocumentID
}

// TestHTTP_InitializeAndToolsList covers (a)+(b): handshake succeeds and
// tools/list advertises exactly the locked 5 tools.
func TestHTTP_InitializeAndToolsList(t *testing.T) {
	ts, keyA, _ := newHTTPServer(t)
	session, ctx, cancel := connectHTTPClient(t, ts.URL, keyA)
	defer cancel()

	if got := session.InitializeResult().ServerInfo.Name; got != "paperviz" {
		t.Errorf("server name = %q, want paperviz", got)
	}

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	want := map[string]bool{"ingest_document": false, "search_documents": false, "get_document": false, "get_figures": false, "get_evidence": false}
	for _, tool := range result.Tools {
		if _, ok := want[tool.Name]; !ok {
			t.Errorf("unexpected tool advertised: %q", tool.Name)
		}
		want[tool.Name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("tool %q not advertised", name)
		}
	}
	if len(result.Tools) != 5 {
		t.Errorf("advertised %d tools, want exactly 5", len(result.Tools))
	}
}

// TestHTTP_RejectsMissingAndBadKeys covers (c): no/wrong key is a 401 that
// never echoes the presented value.
func TestHTTP_RejectsMissingAndBadKeys(t *testing.T) {
	ts, _, keyB := newHTTPServer(t)

	for _, name := range []string{"missing", "wrong"} {
		req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{}`))
		if name == "wrong" {
			req.Header.Set("Authorization", "Bearer "+keyB+"-WRONG")
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		body := make([]byte, 0)
		buf := make([]byte, 256)
		for {
			n, rerr := res.Body.Read(buf)
			body = append(body, buf[:n]...)
			if rerr != nil {
				break
			}
		}
		res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s key: status = %d, want 401", name, res.StatusCode)
		}
		if strings.Contains(string(body), "WRONG") {
			t.Errorf("%s key: response echoes the presented key", name)
		}
	}
}

// TestHTTP_RejectsOversizedBody covers (d): a body beyond the transport
// bound is rejected with 413, never reaching a tool.
func TestHTTP_RejectsOversizedBody(t *testing.T) {
	db, keyA, _ := newHTTPTestDB(t)
	hs := NewHTTPServer(db)
	hs.maxBytes = 1024
	ts := httptest.NewServer(hs.Handler())
	defer ts.Close()

	// Valid JSON larger than the 1KB test bound, so only the transport
	// size check can reject it.
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ingest_document","arguments":{"text":"` + strings.Repeat("x", 2048) + `"}}}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+keyA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Accept", "application/json, text/event-stream")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", res.StatusCode)
	}
}

// TestHTTP_IngestIsExtractionOnly covers (h): ingest over HTTP returns the
// extraction-only shape (document_id + status) and makes no pipeline claim.
func TestHTTP_IngestIsExtractionOnly(t *testing.T) {
	ts, keyA, _ := newHTTPServer(t)
	session, ctx, cancel := connectHTTPClient(t, ts.URL, keyA)
	defer cancel()

	out := callToolText(t, session, ctx, "ingest_document", map[string]any{
		"text": "Quantum entanglement basics. Spin correlations, EPR paradox, Bell inequalities.",
	})
	if !strings.Contains(out, `"document_id"`) {
		t.Errorf("result = %s, want document_id", out)
	}
	for _, leaked := range []string{"simplified_text", "original_text", "tokens", "llm"} {
		if strings.Contains(out, leaked) {
			t.Errorf("result = %s, unexpected field %q", out, leaked)
		}
	}
}

// TestHTTP_OwnerIsolation covers (f): user B cannot read user A's documents
// through any of the three read tools, and B's search never surfaces A's
// documents. Existence of A's document is indistinguishable from absence.
func TestHTTP_OwnerIsolation(t *testing.T) {
	ts, keyA, keyB := newHTTPServer(t)
	sessionA, ctxA, cancelA := connectHTTPClient(t, ts.URL, keyA)
	defer cancelA()
	docID := ingestText(t, sessionA, ctxA, "Zebra migration patterns across the Serengeti savanna.")

	sessionB, ctxB, cancelB := connectHTTPClient(t, ts.URL, keyB)
	defer cancelB()

	// get_document must fail as if the document did not exist.
	got := callToolText(t, sessionB, ctxB, "get_document", map[string]any{"document_id": docID})
	if strings.Contains(got, "Zebra migration patterns") {
		t.Errorf("B read A's document via get_document: %s", got)
	}
	if !strings.Contains(got, "not found") {
		t.Errorf("B get_document error = %s, want not-found", got)
	}

	// get_figures / get_evidence must not return A's data either.
	for _, tool := range []string{"get_figures", "get_evidence"} {
		out := callToolText(t, sessionB, ctxB, tool, map[string]any{"document_id": docID})
		if strings.Contains(out, "Zebra migration patterns") {
			t.Errorf("B read A's document via %s: %s", tool, out)
		}
	}

	// search_documents must never surface A's document.
	search := callToolText(t, sessionB, ctxB, "search_documents", map[string]any{"query": "Zebra"})
	if strings.Contains(search, docID) {
		t.Errorf("B search surfaced A's document: %s", search)
	}

	// And the owner still reads their own document.
	own := callToolText(t, sessionA, ctxA, "get_document", map[string]any{"document_id": docID})
	if !strings.Contains(own, "Zebra migration patterns") {
		t.Errorf("owner read own document: %s", own)
	}
}

// TestHTTP_RateLimitTrips covers (e): the analyze budget (5/min burst 5)
// trips on the 6th ingest under the same identity.
func TestHTTP_RateLimitTrips(t *testing.T) {
	ts, keyA, _ := newHTTPServer(t)
	session, ctx, cancel := connectHTTPClient(t, ts.URL, keyA)
	defer cancel()

	limited := false
	for i := 0; i < 6 && !limited; i++ {
		out := callToolText(t, session, ctx, "ingest_document", map[string]any{
			"text": "A short paper body, just long enough to ingest deterministically.",
		})
		limited = strings.Contains(out, "rate limit exceeded")
	}
	if !limited {
		t.Error("6 rapid ingests never tripped the analyze rate limit (burst 5)")
	}
}
