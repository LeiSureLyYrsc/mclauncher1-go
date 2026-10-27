package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/bmclapi"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/config"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/launch"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	c := bmclapi.NewClient()
	m, err := c.FetchManifest(ctx)
	if err != nil {
		panic(err)
	}
	id := "1.16.5"
	info, ok := bmclapi.FindVersionInfo(m, id)
	if !ok {
		panic("no " + id)
	}
	cfg := config.Default()
	cfg.GameDir = "/tmp/mc-smoke-game"
	fmt.Println("installing", id, "…")
	vj, err := c.InstallVersion(ctx, info, bmclapi.InstallOptions{
		GameDir: cfg.GameDir,
		Version: id,
		Progress: func(msg string) {
			fmt.Println(" ", msg)
		},
	})
	if err != nil {
		fmt.Println("install err:", err)
		os.Exit(1)
	}
	fmt.Println("installed", vj.ID, "main", vj.MainClass)

	cmd, err := launch.Launch(ctx, &cfg, id, func(msg string) { fmt.Println(" ", msg) })
	fmt.Println("launch:", err)
	if cmd != nil && cmd.Process != nil {
		fmt.Println("pid", cmd.Process.Pid)
		_ = cmd.Process.Kill()
	}
}
