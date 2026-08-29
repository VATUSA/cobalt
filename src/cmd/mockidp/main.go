// mockidp is a throwaway stand-in for VATSIM Connect, used only by the
// docker-compose.test.yml integration stack. It implements just the three
// endpoints vatsim/connect.go calls (authorize, token, user) against a
// handful of canned profiles selected by the mock_scenario query param on
// /oauth/authorize, so login.hurl et al can drive cobalt's real OAuth code
// path without a human or a dependency on the real VATSIM sandbox.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"vatusa-cobalt/vatsim"
)

// scenarios maps the mock_scenario query param to a canned VATSIM Connect
// profile. Keep CIDs stable — tests/fixtures/seed.sql seeds `user` rows for
// exactly these CIDs, since nothing in cobalt's login path creates one.
var scenarios = map[string]vatsim.ConnectUserDataData{
	"active": {
		CID: "900001",
		Personal: vatsim.ConnectUserDataPersonal{
			NameFirst: "Mock",
			NameLast:  "Active",
			Email:     "mock-active@example.test",
		},
		Vatsim: vatsim.ConnectUserDataVatsim{
			Rating:      vatsim.ConnectUserDataRating{Id: 2, Long: "Tower", Short: "S2"},
			PilotRating: vatsim.ConnectUserDataRating{Id: 0, Long: "Basic", Short: "PPL"},
			Division:    vatsim.ConnectUserDataUnit{Id: "USA", Name: "United States"},
			Region:      vatsim.ConnectUserDataUnit{Id: "AMAS", Name: "Americas"},
			SubDivision: vatsim.ConnectUserDataUnit{Id: "", Name: ""},
		},
	},
	"suspended": {
		CID: "900002",
		Personal: vatsim.ConnectUserDataPersonal{
			NameFirst: "Mock",
			NameLast:  "Suspended",
			Email:     "mock-suspended@example.test",
		},
		Vatsim: vatsim.ConnectUserDataVatsim{
			// 0 == config.RatingSuspended; see login.go's inactive/suspended check.
			Rating:      vatsim.ConnectUserDataRating{Id: 0, Long: "Suspended", Short: "SUS"},
			PilotRating: vatsim.ConnectUserDataRating{Id: 0, Long: "Basic", Short: "PPL"},
			Division:    vatsim.ConnectUserDataUnit{Id: "USA", Name: "United States"},
			Region:      vatsim.ConnectUserDataUnit{Id: "AMAS", Name: "Americas"},
			SubDivision: vatsim.ConnectUserDataUnit{Id: "", Name: ""},
		},
	},
}

type pendingCode struct {
	scenario    string
	redirectURI string
}

type issuedToken struct {
	scenario string
}

type store struct {
	mu sync.Mutex
	// defaultScenario is which canned profile /oauth/authorize uses. It's
	// set via POST /mock/set-scenario rather than a query param on
	// /oauth/authorize, because cobalt's GetLogin (endpoints/login.go)
	// builds that URL itself and never forwards a caller-supplied
	// mock_scenario through it -- the only channel a test has to steer which
	// profile a /login redirect will land on is calling the mock IdP
	// directly first. Tests run serially against one mock-idp instance, so
	// this single shared value is enough.
	defaultScenario string
	codes           map[string]pendingCode
	tokens          map[string]issuedToken
}

func newStore() *store {
	return &store{
		defaultScenario: "active",
		codes:           map[string]pendingCode{},
		tokens:          map[string]issuedToken{},
	}
}

func (s *store) setScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Scenario string `json:"scenario"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
	if _, ok := scenarios[body.Scenario]; !ok {
		http.Error(w, "unknown scenario", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.defaultScenario = body.Scenario
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("mockidp: failed to generate random token: %v", err)
	}
	return hex.EncodeToString(b)
}

func (s *store) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	if redirectURI == "" {
		http.Error(w, "missing redirect_uri", http.StatusBadRequest)
		return
	}
	scenario := q.Get("mock_scenario")
	if scenario == "" {
		s.mu.Lock()
		scenario = s.defaultScenario
		s.mu.Unlock()
	}
	if _, ok := scenarios[scenario]; !ok {
		http.Error(w, "unknown mock_scenario", http.StatusBadRequest)
		return
	}

	code := randomHex(16)
	s.mu.Lock()
	s.codes[code] = pendingCode{scenario: scenario, redirectURI: redirectURI}
	s.mu.Unlock()

	dest, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	destQuery := dest.Query()
	destQuery.Set("code", code)
	if state := q.Get("state"); state != "" {
		destQuery.Set("state", state)
	}
	dest.RawQuery = destQuery.Encode()
	http.Redirect(w, r, dest.String(), http.StatusFound)
}

func (s *store) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form body", http.StatusBadRequest)
		return
	}
	if r.FormValue("client_id") != os.Getenv("MOCK_IDP_CLIENT_ID") ||
		r.FormValue("client_secret") != os.Getenv("MOCK_IDP_CLIENT_SECRET") {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	code := r.FormValue("code")
	s.mu.Lock()
	pending, ok := s.codes[code]
	if ok {
		delete(s.codes, code)
	}
	s.mu.Unlock()
	if !ok {
		http.Error(w, "invalid or already-used code", http.StatusBadRequest)
		return
	}

	accessToken := randomHex(16)
	s.mu.Lock()
	s.tokens[accessToken] = issuedToken{scenario: pending.scenario}
	s.mu.Unlock()

	writeJSON(w, vatsim.ConnectAccessToken{
		Scopes:      []string{"full_name", "email", "vatsim_details"},
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		AccessToken: accessToken,
	})
}

func (s *store) user(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(prefix):]

	s.mu.Lock()
	issued, ok := s.tokens[token]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	writeJSON(w, vatsim.ConnectUserData{Data: scenarios[issued.scenario]})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("mockidp: failed to encode response: %v", err)
	}
}

func main() {
	s := newStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", s.authorize)
	mux.HandleFunc("/oauth/token", s.token)
	mux.HandleFunc("/api/user", s.user)
	mux.HandleFunc("/mock/set-scenario", s.setScenario)

	addr := "0.0.0.0:9100"
	log.Printf("mockidp listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("mockidp: server failed: %v", err)
	}
}
