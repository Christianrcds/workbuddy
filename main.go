package main

import (
	"embed"
	"workbuddy/cmd"
)

//go:embed migrations
var migrationsFS embed.FS

func main() {
	cmd.SetMigrations(migrationsFS)
	cmd.Execute()
}
