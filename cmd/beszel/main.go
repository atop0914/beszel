package main

import (
	"flag"
	"log"
	"os"

	"github.com/atop0914/beszel/internal/collector"
	"github.com/atop0914/beszel/internal/config"
	"github.com/atop0914/beszel/internal/store"
	"github.com/atop0914/beszel/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		serveCmd.Parse(os.Args[2:])
		cfg, err := config.Load("beszel.yaml")
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		db, err := store.New(cfg.DBPath)
		if err != nil {
			log.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()
		web.Run(cfg, db)

	case "collect-once":
		collectOnceCmd := flag.NewFlagSet("collect-once", flag.ExitOnError)
		collectOnceCmd.Parse(os.Args[2:])
		cfg, err := config.Load("beszel.yaml")
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		db, err := store.New(cfg.DBPath)
		if err != nil {
			log.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()
		col, err := collector.NewSystemCollector()
		if err != nil {
			log.Fatalf("create collector: %v", err)
		}
		m, err := col.Collect()
		if err != nil {
			log.Fatalf("collection failed: %v", err)
		}
		if err := db.InsertMetrics(m); err != nil {
			log.Fatalf("insert failed: %v", err)
		}
		log.Println("collection complete")

	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	log.Println("Usage: beszel [serve|collect-once]")
}
