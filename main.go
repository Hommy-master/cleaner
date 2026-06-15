package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"cleaner/internal/cleaner"
	"cleaner/internal/config"
	"cleaner/internal/util"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get working directory: %v\n", err)
		os.Exit(1)
	}

	if err := util.InitLogger(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	logger := util.GetLogger()

	cfgPath := *configPath
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(workDir, cfgPath)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Printf("ERROR: load config: %v", err)
		os.Exit(1)
	}

	logger.Printf("cleaner started, interval=%ds, config=%s", cfg.Interval, cfgPath)

	c := cleaner.New(cfg, logger)
	c.Run()

	ticker := time.NewTicker(cfg.IntervalDuration())
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			c.Run()
		case sig := <-sigCh:
			logger.Printf("received signal %v, shutting down", sig)
			return
		}
	}
}
