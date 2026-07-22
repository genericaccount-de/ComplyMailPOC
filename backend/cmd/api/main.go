package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/genericaccount-de/comply-mail-poc/backend/internal/config"
	"github.com/genericaccount-de/comply-mail-poc/backend/internal/llm"
	"github.com/genericaccount-de/comply-mail-poc/backend/internal/rest/handler"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	llmClient := llm.New(llm.Config{
		BaseURL: cfg.LLM.BaseURL,
		APIKey:  cfg.LLM.APIKey,
		Model:   cfg.LLM.Model,
		Timeout: time.Duration(cfg.LLM.Timeout),
	})

	styleCheckHandler := handler.NewStyleCheck(llmClient, "")
	scanHandler := handler.NewScan(llmClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	r.Post("/check-style-email", styleCheckHandler.ServeHTTP)
	r.Post("/scan-outbound-email", scanHandler.ServeHTTP)

	log.Printf("ComplyMail API listening on %s (model=%s)", cfg.Server.ListenAddr, llmClient.Model())
	if err := http.ListenAndServe(cfg.Server.ListenAddr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
