package main

import (
	"log"
	"net/http"
	"os"

	"edj.cash/edj/internal/api"
	"edj.cash/edj/internal/chain"
	"edj.cash/edj/internal/store"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dataFile := os.Getenv("EDJ_DATA_FILE")
	if dataFile == "" {
		dataFile = "data/edj.json"
	}
	memory, signer, err := store.NewPersistent(dataFile)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	s := &api.Server{
		Chain: chain.BCH{
			ReceiveAddress: os.Getenv("EDJ_RECEIVE_ADDRESS"),
			TokenAddress:   os.Getenv("EDJ_TOKEN_ADDRESS"),
		},
		Store:  memory,
		Signer: signer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.Health)
	mux.HandleFunc("GET /api/config", s.Config)
	mux.HandleFunc("POST /api/auth/google", s.GoogleAuth)
	mux.HandleFunc("GET /api/me", s.Me)
	mux.HandleFunc("GET /api/portfolio", s.RequireUser(s.Portfolio))
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

	log.Printf("edj listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
