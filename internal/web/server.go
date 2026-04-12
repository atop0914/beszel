package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atop0914/beszel/internal/collector"
	"github.com/atop0914/beszel/internal/config"
	"github.com/atop0914/beszel/internal/store"
)

type Server struct {
	cfg           *config.Config
	db            *store.Store
	collector     *collector.SystemCollector
	dockerCollector *collector.DockerCollector
}

func Run(cfg *config.Config, db *store.Store) {
	col, err := collector.NewSystemCollector()
	if err != nil {
		log.Fatalf("create collector: %v", err)
	}
	s := &Server{cfg: cfg, db: db, collector: col}

	// Initialize Docker collector if enabled
	if cfg.DockerEnabled {
		dc, err := collector.NewDockerCollector()
		if err != nil {
			log.Printf("docker collector init failed (docker may not be available): %v", err)
		} else {
			s.dockerCollector = dc
			log.Println("docker collector initialized")
		}
	}

	// Background collection loop
	go s.collectionLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/metrics/history", s.handleMetricsHistory)
	mux.HandleFunc("/api/containers", s.handleContainers)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("beszel listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func (s *Server) collectionLoop() {
	ticker := time.NewTicker(s.cfg.CollectionInterval)
	defer ticker.Stop()
	pruneTicker := time.NewTicker(time.Hour)
	defer pruneTicker.Stop()

	// Prune old data on startup
	if err := s.db.Prune(7 * 24 * time.Hour); err != nil {
		log.Printf("prune error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			m, err := s.collector.Collect()
			if err != nil {
				log.Printf("collect error: %v", err)
				continue
			}
			if err := s.db.InsertMetrics(m); err != nil {
				log.Printf("insert metrics error: %v", err)
			}

			// Collect Docker container stats if available
			if s.dockerCollector != nil {
				containers, err := s.dockerCollector.CollectContainerStats()
				if err != nil {
					log.Printf("docker collect error: %v", err)
				} else {
					for _, c := range containers {
						if err := s.db.InsertContainerMetrics(c); err != nil {
							log.Printf("insert container metrics error: %v", err)
						}
					}
				}
			}
		case <-pruneTicker.C:
			if err := s.db.Prune(7 * 24 * time.Hour); err != nil {
				log.Printf("prune error: %v", err)
			}
		}
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	m, err := s.db.LatestMetrics()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func (s *Server) handleMetricsHistory(w http.ResponseWriter, r *http.Request) {
	hours := 24
	fmt.Sscanf(r.URL.Query().Get("hours"), "%d", &hours)
	metrics, err := s.db.MetricsHistory(hours)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := s.db.LatestContainers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
