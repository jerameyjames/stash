package main

import (
	"context"
	"log"
	"os"

	"github.com/alash3al/stash/internal/bootstrap"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "stash",
		Usage: "Stash - Memory layer for AI applications",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			bc, err := bootstrap.New(ctx)
			if err != nil {
				return ctx, err
			}
			cmd.Metadata["bootstrapCtx"] = bc
			return ctx, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			if bc, ok := cmd.Metadata["bootstrapCtx"].(*bootstrap.Context); ok {
				return bc.Close()
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   "env",
				Usage:  "Show environment variables and configuration",
				Action: EnvCmd,
			},
			{
				Name:    "remember",
				Aliases: []string{"add"},
				Usage:   "Store a memory",
				Action:  rememberCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "context",
						Usage: "Context for the memory (e.g. work, personal)",
					},
					&cli.StringFlag{
						Name:  "metadata",
						Usage: "JSON metadata for the memory",
					},
				},
			},
			{
				Name:    "recall",
				Aliases: []string{"search"},
				Usage:   "Search for relevant memories",
				Action:  recallCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "context",
						Usage: "Context to search (omit to search all)",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of results",
						Value: 10,
					},
				},
			},
			{
				Name:   "forget",
				Usage:  "Forget a memory matching a description",
				Action: forgetCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "context",
						Usage: "Context to search in",
					},
				},
			},
			{
				Name:   "purge",
				Usage:  "Hard-delete a memory by ID",
				Action: purgeCmd,
			},
			{
				Name:   "reflect",
				Usage:  "Introspect memory state and generate report",
				Action: reflectCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "context",
						Usage: "Context to reflect on (optional)",
					},
				},
			},
			{
				Name:   "contradict",
				Usage:  "Find contradictions in memories",
				Action: contradictCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "context",
						Usage: "Context to check (optional)",
					},
				},
			},
			{
				Name:  "mcp",
				Usage: "MCP server for agent integration",
				Commands: []*cli.Command{
					{
						Name:   "serve",
						Usage:  "Start MCP server over SSE",
						Action: mcpServeCmd,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "host",
								Usage: "Server host",
								Value: "0.0.0.0",
							},
							&cli.StringFlag{
								Name:  "port",
								Usage: "Server port",
								Value: "8080",
							},
						},
					},
					{
						Name:   "execute",
						Usage:  "Start MCP server over stdio",
						Action: mcpExecuteCmd,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
