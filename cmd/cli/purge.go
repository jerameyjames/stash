package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alash3al/stash/internal/bootstrap"
	"github.com/urfave/cli/v3"
)

func purgeCmd(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("memory ID argument is required")
	}

	id := args.First()
	if id == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	bc, ok := cmd.Root().Metadata["bootstrapCtx"].(*bootstrap.Context)
	if !ok {
		return fmt.Errorf("bootstrap context not available")
	}

	if err := bc.Brain.Purge(ctx, id); err != nil {
		return err
	}

	output := map[string]string{
		"message": "Memory purged successfully",
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Println(string(jsonOutput))
	return nil
}
