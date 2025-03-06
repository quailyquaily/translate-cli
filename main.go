package main

import (
	"context"

	"github.com/quailyquaily/translate-cli/cmd"
)

var (
	Version = "0.0.1"
)

type contextKey string

const (
	versionKey contextKey = "version"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, versionKey, Version)
	cmd.ExecuteContext(ctx)
}
