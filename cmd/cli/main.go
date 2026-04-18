package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alash3al/stash/internal/bootstrap"
	"github.com/alash3al/stash/internal/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "stash",
		Usage: "Stash - Memory layer for AI applications",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Run the Stash service",
				Action: runCmd,
			},
			{
				Name:   "env",
				Usage:  "Show environment variables and configuration",
				Action: envCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Value:   ".env",
						Usage:   "Config file to load",
					},
				},
			},
			{
				Name:   "check",
				Usage:  "Check configuration and bootstrap",
				Action: checkCmd,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runCmd(ctx context.Context, cmd *cli.Command) error {
	bootstrapCtx := bootstrap.MustNew(ctx)
	defer func() {
		if err := bootstrapCtx.Close(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}()

	fmt.Printf("Stash initialized successfully!\n")
	fmt.Printf("  Store driver: %s\n", bootstrapCtx.Config.StoreDriver)
	fmt.Printf("  Embedder driver: %s\n", bootstrapCtx.Config.EmbedderDriver)
	fmt.Printf("  Vector dimension: %d\n", bootstrapCtx.Config.VectorDim)
	fmt.Printf("  Frame TTL: %v\n", bootstrapCtx.Config.FrameTTL)
	fmt.Printf("\nReady to use memory operations.\n")
	fmt.Printf("Press Ctrl+C to exit.\n")

	<-ctx.Done()
	fmt.Printf("\nShutting down...\n")
	return nil
}

func envCmd(ctx context.Context, cmd *cli.Command) error {
	configFile := cmd.String("file")

	// Use same logic as bootstrap to determine config file
	filename := os.Getenv("STASHCONFIG")
	if filename == "" {
		filename = configFile
	}

	fmt.Printf("Configuration file: %s\n", filename)
	fmt.Printf("STASHCONFIG env var: %s\n", os.Getenv("STASHCONFIG"))
	fmt.Println()

	cfg, err := config.NewFromFile(filename)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	fmt.Println("=== Loaded Configuration ===")
	fmt.Printf("Store Driver: %s\n", cfg.StoreDriver)
	fmt.Printf("Store DSN: %s\n", maskDSN(cfg.StoreDSN))
	fmt.Printf("Vector Dimension: %d\n", cfg.VectorDim)
	fmt.Printf("Max Result Size: %d\n", cfg.MaxResultSize)
	fmt.Printf("Embedder Driver: %s\n", cfg.EmbedderDriver)
	fmt.Printf("OpenAI API Key: %s\n", maskAPIKey(cfg.OpenAIAPIKey))
	fmt.Printf("OpenAI Base URL: %s\n", cfg.OpenAIBaseURL)
	fmt.Printf("Embedding Model: %s\n", cfg.EmbeddingModel)
	fmt.Printf("Frame TTL: %v\n", cfg.FrameTTL)
	fmt.Printf("HTTP Addr: %s\n", cfg.HTTPAddr)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)

	fmt.Println("\n=== Bootstrap Test ===")
	bootstrapCtx, err := bootstrap.New(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}
	defer bootstrapCtx.Close()

	fmt.Println("Bootstrap successful!")
	fmt.Printf("Store initialized: %v\n", bootstrapCtx.Store != nil)
	fmt.Printf("Embedder initialized: %v\n", bootstrapCtx.Embedder != nil)
	fmt.Printf("Memory initialized: %v\n", bootstrapCtx.Memory != nil)

	return nil
}

func checkCmd(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Checking configuration and bootstrap...")
	
	bootstrapCtx, err := bootstrap.New(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap check failed: %w", err)
	}
	defer bootstrapCtx.Close()

	fmt.Println("✓ Configuration loaded successfully")
	fmt.Println("✓ Store initialized successfully")
	fmt.Println("✓ Embedder initialized successfully")
	fmt.Println("✓ Memory layer initialized successfully")
	fmt.Println("\nAll checks passed!")
	
	return nil
}

func maskDSN(dsn string) string {
	if len(dsn) > 50 {
		return dsn[:20] + "..." + dsn[len(dsn)-20:]
	}
	return dsn
}

func maskAPIKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}