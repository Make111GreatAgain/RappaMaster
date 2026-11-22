package main

import (
	"BHLayer2Node/Database"
	"BHLayer2Node/paradigm"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "application JSON config path (requires explicit database settings)")
	apply := flag.Bool("apply", false, "delete eligible rows (default: preview only)")
	batches := flag.Int("max-batches", 10, "maximum 1000-row batches per apply run")
	flag.Parse()
	if *batches < 1 || flag.NArg() != 0 {
		log.Fatal("expected a positive max-batches and no positional arguments")
	}
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	var config paradigm.BHLayer2NodeConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal("cannot decode application config")
	}
	file.Close()
	if config.Database == nil || config.Database.Host == "" || config.Database.Dbname == "" {
		log.Fatal("explicit database host and dbname are required")
	}
	config.DEBUG = false
	db, err := Database.NewDatabaseService(&config, nil)
	if err != nil {
		log.Fatal("cannot connect to the configured database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cutoff := time.Now().AddDate(0, 0, -30).Truncate(time.Second)
	fmt.Printf("database=%s cutoff=%s\n", config.Database.Dbname, cutoff.Format(time.RFC3339))
	if !*apply {
		p, err := db.PreviewEmptyEpochCleanup(ctx, cutoff)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("preview eligible=%d data_bytes=%d index_bytes=%d free_bytes=%d\n",
			p.Eligible, p.DataBytes, p.IndexBytes, p.FreeBytes)
		return
	}
	var deleted int64
	for i := 0; i < *batches; i++ {
		n, err := db.DeleteEmptyEpochBatch(ctx, cutoff)
		if err != nil {
			log.Fatalf("cleanup stopped after %d committed deletions: %v", deleted, err)
		}
		deleted += n
		if n < 1000 {
			break
		}
	}
	fmt.Printf("deleted=%d\n", deleted)
}
