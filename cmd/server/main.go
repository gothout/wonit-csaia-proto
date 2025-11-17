package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whatsapp-ia-integrator/internal/chatvolt"
	"whatsapp-ia-integrator/internal/config"
	"whatsapp-ia-integrator/internal/csa"
	"whatsapp-ia-integrator/internal/queue"
	"whatsapp-ia-integrator/internal/session"
	"whatsapp-ia-integrator/internal/whatsapp"
)

func main() {
	cfgPath := flag.String("config", "config.json", "caminho do arquivo de configuração")
	workers := flag.Int("workers", 3, "quantidade de workers para envio")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("erro carregando config: %v", err)
	}

	csaClient := csa.NewClient(cfg.CSA)
	chatvoltClient := chatvolt.NewClient(cfg.IA.Chatvolt)
	sessionManager := session.NewManager(10 * time.Minute)
	jobManager := queue.NewJobManager()

	outbox := queue.NewOutbox(csaClient, *workers, jobManager)
	outbox.Start()

	handler := whatsapp.NewHandler(chatvoltClient, sessionManager, outbox, jobManager)

	mux := http.NewServeMux()
	mux.Handle("/whatsapp/webhook", handler)
	mux.Handle("/jobs/", queue.NewJobStatusHandler(jobManager))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("servidor escutando em %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro no servidor: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("encerrando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("erro encerrando servidor: %v", err)
	}

	outbox.Stop()
}
