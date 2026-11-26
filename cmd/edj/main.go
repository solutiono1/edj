package main

import (
	"log"
	"net/http"
	"os"

	"edj.cash/edj/internal/api"
	"edj.cash/edj/internal/chain"
	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		if port := os.Getenv("PORT"); port != "" {
			addr = ":" + port
		} else {
			addr = ":8080"
		}
	}
	var storeImpl store.Store
	var signer *core.Signer
	var err error
	if supabaseURL := os.Getenv("SUPABASE_URL"); supabaseURL != "" {
		key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		if key == "" {
			log.Fatal("SUPABASE_SERVICE_ROLE_KEY is required when SUPABASE_URL is set")
		}
		storeImpl, signer, err = store.NewSupabase(supabaseURL, key)
	} else {
		dataFile := os.Getenv("EDJ_DATA_FILE")
		if dataFile == "" {
			dataFile = "data/edj.json"
		}
		var memory *store.Memory
		memory, signer, err = store.NewPersistent(dataFile)
		storeImpl = memory
	}
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	s := &api.Server{
		Chain: chain.BCH{
			ReceiveAddress: os.Getenv("EDJ_RECEIVE_ADDRESS"),
			TokenAddress:   os.Getenv("EDJ_TOKEN_ADDRESS"),
		},
		Store:  storeImpl,
		Signer: signer,
	}
	if host := os.Getenv("EDJ_ELECTRUM_HOST"); host != "" {
		s.Electrum = &chain.Electrum{Host: host}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.Health)
	mux.HandleFunc("GET /api/config", s.Config)
	mux.HandleFunc("POST /api/auth/google", s.GoogleAuth)
	mux.HandleFunc("POST /api/auth/signup", s.EmailSignup)
	mux.HandleFunc("POST /api/auth/login", s.EmailLogin)
	mux.HandleFunc("GET /api/me", s.Me)
	mux.HandleFunc("GET /api/portfolio", s.RequireUser(s.Portfolio))
	mux.HandleFunc("GET /api/wallet/utxos", s.WalletUtxos)
	mux.HandleFunc("POST /api/wallet/broadcast", s.WalletBroadcast)
	mux.HandleFunc("POST /api/logout", s.Logout)
	mux.HandleFunc("GET /api/notifications", s.RequireUser(s.Notifications))
	mux.HandleFunc("POST /api/assistant", s.Assistant)
	mux.HandleFunc("POST /api/links", s.RequireUser(s.CreateLink))
	mux.HandleFunc("GET /api/links/{id}", s.GetLink)
	mux.HandleFunc("POST /api/payroll/upload", s.RequireUser(s.UploadPayroll))
	mux.HandleFunc("GET /api/payroll/{id}", s.RequireUser(s.GetPayroll))
	mux.HandleFunc("POST /api/disclose", s.RequireUser(s.CreateProof))
	mux.HandleFunc("GET /api/proofs/{id}", s.GetProof)

	static := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", static))

	serve := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "web/"+name)
		}
	}
	mux.HandleFunc("GET /", serve("index.html"))
	mux.HandleFunc("GET /login", serve("login.html"))
	mux.HandleFunc("GET /signup", serve("signup.html"))
	mux.HandleFunc("GET /onboarding", serve("onboarding.html"))
	for _, p := range []string{"/portfolio", "/links", "/payroll", "/proofs", "/transactions", "/wallet", "/notifications", "/profile", "/settings"} {
		mux.HandleFunc("GET "+p, serve("app.html"))
	}
	mux.HandleFunc("GET /link/new", serve("link/new.html"))
	mux.HandleFunc("GET /payroll/new", serve("payroll/new.html"))
	mux.HandleFunc("GET /proof/new", serve("proof/new.html"))
	mux.HandleFunc("GET /proof/{id}", serve("proof/view.html"))
	mux.HandleFunc("GET /l/{id}", serve("l/claim.html"))
	mux.HandleFunc("GET /terms", serve("legal.html"))
	mux.HandleFunc("GET /privacy", serve("legal.html"))
	mux.HandleFunc("GET /delete", serve("legal.html"))
	mux.HandleFunc("GET /docs", serve("docs.html"))
	mux.HandleFunc("GET /faq", serve("index.html"))
	mux.HandleFunc("GET /support", serve("index.html"))

	log.Printf("pagor listening on %s", addr)
	origin := os.Getenv("EDJ_PUBLIC_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	log.Fatal(http.ListenAndServe(addr, withCORS(origin, mux)))
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
