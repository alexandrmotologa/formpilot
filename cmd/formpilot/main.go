package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/dedup"
	"github.com/alexandrmotologa/formpilot/internal/pipeline"
	"github.com/alexandrmotologa/formpilot/internal/receiver"
	"github.com/alexandrmotologa/formpilot/internal/server"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

var (
	Version   = "1.0.0"
	BuildDate = "2026-11-17"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "ingest":
		cmdIngest(os.Args[2:])
	case "simulate":
		cmdSimulate(os.Args[2:])
	case "status":
		cmdStatus(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("formpilot v%s (built %s, %s/%s)\n", Version, BuildDate, runtime.GOOS, runtime.GOARCH)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`formpilot - Autonomous Unstructured-to-CRM Ingestion Engine

Usage:
  formpilot <command> [flags]

Commands:
  serve       Start the daemon, webhook receivers, and embedded Sales Studio
  ingest      Process an email (.eml), text, or PDF RFQ file directly
  simulate    Inject a simulated enterprise inbound lead into the pipeline
  status      Display database metrics, memory usage, and CRM connectivity
  version     Display application version information

Flags for 'serve':
  -config string    Path to configuration YAML file (default: examples/formpilot.yaml)
  -port int         Port to listen on (default: 9393)
  -host string      Host to bind to (default: 0.0.0.0)

Flags for 'ingest':
  -file string      Path to .eml, .txt, or .pdf file to ingest (required)
  -config string    Path to configuration YAML file

Flags for 'simulate':
  -preset string    Simulation preset: fintech, saas, hipaa, freemail, spam (default: fintech)
  -config string    Path to configuration YAML file`)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "examples/formpilot.yaml", "Path to config file")
	port := fs.Int("port", 0, "Server port")
	host := fs.String("host", "", "Server host")
	_ = fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *host != "" {
		cfg.Server.Host = *host
	}

	db, err := storage.Open(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	cache, err := dedup.NewAccountCache(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing account cache: %v\n", err)
		os.Exit(1)
	}

	sse := server.NewSSEBroadcaster()
	pipe := pipeline.NewPipeline(cfg, db, cache, sse)

	srv := server.NewServer(cfg, db, sse)
	srv.SetPipeline(pipe)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("=================================================================")
		fmt.Printf(" formpilot Ingestion Daemon v%s running on http://%s\n", Version, addr)
		fmt.Printf(" Embedded Sales Ops Cockpit: http://localhost:%d\n", cfg.Server.Port)
		fmt.Printf(" Mode: CRM [%s] | Extraction [%s] | DB [%s]\n", cfg.CRM.Mode, cfg.Extraction.Provider, cfg.Database.Path)
		fmt.Println("=================================================================")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	<-sigChan
	fmt.Println("\nReceived shutdown signal. Stopping formpilot gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = httpServer.Shutdown(ctx)
	fmt.Println("formpilot stopped cleanly.")
}

func cmdIngest(args []string) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	filePath := fs.String("file", "", "Path to email or PDF file")
	configPath := fs.String("config", "examples/formpilot.yaml", "Path to config file")
	_ = fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --file is required for ingest")
		os.Exit(1)
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *filePath, err)
		os.Exit(1)
	}

	cfg, _ := config.LoadConfig(*configPath)
	db, _ := storage.Open(cfg.Database.Path)
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()

	cache, _ := dedup.NewAccountCache(db)
	sse := server.NewSSEBroadcaster()
	pipe := pipeline.NewPipeline(cfg, db, cache, sse)

	payload, err := receiver.ParseFileToInbound(*filePath, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	lead, err := pipe.Process(context.Background(), payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline execution error: %v\n", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(lead, "", "  ")
	fmt.Println(string(out))
	fmt.Printf("\n[SUCCESS] Ingested in %dms | Company: %s | Score: %d | Tier: %s\n", lead.LatencyMs, lead.CompanyName, lead.ICPScore, lead.QualificationTier)
}

func cmdSimulate(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ExitOnError)
	preset := fs.String("preset", "fintech", "Simulation preset: fintech, saas, hipaa, freemail, spam")
	configPath := fs.String("config", "examples/formpilot.yaml", "Path to config file")
	_ = fs.Parse(args)

	cfg, _ := config.LoadConfig(*configPath)
	db, _ := storage.Open(cfg.Database.Path)
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()

	cache, _ := dedup.NewAccountCache(db)
	sse := server.NewSSEBroadcaster()
	pipe := pipeline.NewPipeline(cfg, db, cache, sse)

	var payload *receiver.InboundPayload

	switch *preset {
	case "saas":
		payload = &receiver.InboundPayload{
			Source:    "simulation_saas",
			Sender:    "Elena Rostova <elena@cloudscale-analytics.io>",
			Recipient: "leads@formpilot.io",
			Subject:   "Demo request for 45 seats - CloudScale Analytics",
			Body: `Hello, My name is Elena Rostova and I head the Product Operations team at CloudScale Analytics (cloudscale-analytics.io).
We are looking for HubSpot CRM integration for 45 seats.
Our budget is around $35,000 annually. Timeline is 2 months.
Phone: +1 650 443 9081`,
		}

	case "hipaa":
		payload = &receiver.InboundPayload{
			Source:    "simulation_hipaa",
			Sender:    "Dr. Jonathan Hayes <jhayes@stjudehealthsystem.org>",
			Recipient: "rfq@formpilot.io",
			Subject:   "HIPAA Compliant Intake Automation - St Jude Health System",
			Body: `Dear Enterprise Inquiries,
I am Dr. Jonathan Hayes, Chief Medical Information Officer at St Jude Health System (stjudehealthsystem.org).
Scope: 500 clinical coordinators.
Compliance: Strict HIPAA BAA required. Integration with Salesforce Health Cloud.
Budget: Board approved $300,000 capital expenditure.
Urgency: P0 immediate deployment.`,
		}

	case "freemail":
		payload = &receiver.InboundPayload{
			Source:    "simulation_freemail",
			Sender:    "alex.dev99@gmail.com",
			Recipient: "info@formpilot.io",
			Subject:   "Question about features",
			Body:      "Hey, just testing if this works for a personal project with 2 friends. Thanks, Alex",
		}

	case "spam":
		payload = &receiver.InboundPayload{
			Source:    "simulation_spam",
			Sender:    "blast@tempmail.com",
			Recipient: "contact@formpilot.io",
			Subject:   "Cheap SEO 5000 Backlinks",
			Body:      "Claim 5000 high PR backlinks for $50. Unsubscribe by replying.",
		}

	default: // "fintech"
		payload = &receiver.InboundPayload{
			Source:    "simulation_fintech",
			Sender:    "Marcus Vance <m.vance@apexclearing.com>",
			Recipient: "sales@formpilot.io",
			Subject:   "Enterprise RFP: Inbound Data Pipeline for 250 FinTech Seats",
			Body: `Hi Sales Team,
I am Marcus Vance, VP of Engineering at Apex Clearing (apexclearing.com).
We need 250 enterprise seats for operations.
Requirements: Automated entity extraction, custom CRM integration, and audit logs.
Security: SOC2 Type II, dedicated VPC, SAML SSO with Okta.
Budget: Allocated $120,000 annual licensing.
Timeline: Immediate Q4 initiative; procurement within 30 days.
Phone: +1 (415) 890-3421`,
		}
	}

	payload.RawPayload = payload.Body

	lead, err := pipe.Process(context.Background(), payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Simulation error: %v\n", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(lead, "", "  ")
	fmt.Println(string(out))
	fmt.Printf("\n[SIMULATION SUCCESS] Latency: %dms | Company: %s | Score: %d/100 | Tier: %s\n", lead.LatencyMs, lead.CompanyName, lead.ICPScore, lead.QualificationTier)
}

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "examples/formpilot.yaml", "Path to config file")
	_ = fs.Parse(args)

	cfg, _ := config.LoadConfig(*configPath)
	db, err := storage.Open(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	stats, err := db.GetStatsSummary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stats: %v\n", err)
		os.Exit(1)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("=================================================================")
	fmt.Printf(" formpilot Status & Health Report (v%s)\n", Version)
	fmt.Println("=================================================================")
	fmt.Printf(" Database:              %s\n", cfg.Database.Path)
	fmt.Printf(" CRM Mode:              %s\n", cfg.CRM.Mode)
	fmt.Printf(" Extraction Provider:   %s\n", cfg.Extraction.Provider)
	fmt.Printf(" Memory Allocated:      %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf(" Memory System Total:   %.2f MB (Target: < 15.00 MB baseline)\n", float64(m.Sys)/1024/1024)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf(" Total Ingested Leads:  %d\n", stats.TotalLeads)
	fmt.Printf(" Enterprise Hot Leads:  %d\n", stats.EnterpriseHotLeads)
	fmt.Printf(" Mid-Market Leads:      %d\n", stats.MidMarketLeads)
	fmt.Printf(" SMB / Self-Serve:      %d\n", stats.SMBLeads)
	fmt.Printf(" Unqualified Leads:     %d\n", stats.UnqualifiedLeads)
	fmt.Printf(" Deduplicated Accounts: %d\n", stats.DeduplicatedAccounts)
	fmt.Printf(" Successful Syncs:      %d\n", stats.SuccessfulSyncs)
	fmt.Printf(" Average Pipeline Time: %.1f ms (Benchmark: < 3000 ms)\n", stats.AverageLatencyMs)
	fmt.Println("=================================================================")
}
