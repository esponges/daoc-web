// gameserver runs the multiplayer game: it serves the viewer, logs players
// in (Discord or an admin password), and runs the zones they play in.
//
//	DAOC_ADMIN_PASSWORD=... go run ./cmd/gameserver
//
// Settings come from flags, the environment, or server/data/server.env
// (KEY=VALUE lines; that directory is gitignored). See the README's
// Multiplayer section.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"daocweb/internal/server/auth"
	"daocweb/internal/server/gameserver"
	"daocweb/internal/server/store"
	"daocweb/internal/server/world"
)

func main() {
	addr := flag.String("addr", env("DAOC_ADDR", "localhost:8777"), "address to listen on (use :8777 for the whole network)")
	webDir := flag.String("web", "web", "the viewer and converted zones")
	dataDir := flag.String("data", filepath.Join("server", "data"), "where the world and player databases live")
	flag.Parse()

	loadEnvFile(filepath.Join(*dataDir, "server.env"))

	worldDB := filepath.Join(*dataDir, "world.db")
	webData := filepath.Join(*webDir, "data")

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatalf("data directory: %v", err)
	}
	st, err := store.Open(filepath.Join(*dataDir, "players.db"))
	if err != nil {
		log.Fatalf("player database: %v", err)
	}
	defer st.Close()

	var cat *world.Catalog
	if _, err := os.Stat(worldDB); err == nil {
		cat, err = world.LoadCatalog(worldDB)
		if err != nil {
			log.Fatalf("world store %s: %v (delete it and run go run ./cmd/dbimport again)", worldDB, err)
		}
	} else {
		log.Printf("no world store at %s: starting with empty zones.\n"+
			"  To populate them: git clone https://github.com/OpenDAoC/OpenDAoC-Database %s\n"+
			"  then: go run ./cmd/dbimport",
			worldDB, filepath.Join(*dataDir, "community"))
		cat = world.EmptyCatalog(webData)
	}

	key := sessionKey(*dataDir)
	a, err := auth.New(auth.Config{
		AdminHash:     os.Getenv("DAOC_ADMIN_HASH"),
		AdminPassword: os.Getenv("DAOC_ADMIN_PASSWORD"),
		DiscordID:     os.Getenv("DAOC_DISCORD_CLIENT_ID"),
		DiscordSecret: os.Getenv("DAOC_DISCORD_CLIENT_SECRET"),
		PublicURL:     env("DAOC_PUBLIC_URL", "http://"+strings.TrimPrefix(*addr, ":")),
		SessionKey:    key,
		Secure:        strings.HasPrefix(os.Getenv("DAOC_PUBLIC_URL"), "https://"),
		Log:           log.Printf,
	}, st)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gameserver:", err)
		if errors.Is(err, auth.ErrNotConfigured) {
			fmt.Fprintln(os.Stderr, "\nFor a local game, set an admin password:\n  DAOC_ADMIN_PASSWORD=choose-one go run ./cmd/gameserver\n"+
				"or put DAOC_ADMIN_PASSWORD=... in "+filepath.Join(*dataDir, "server.env"))
		}
		os.Exit(1)
	}

	w := world.New(world.Config{WebData: webData, Cat: cat, Store: st, Log: log.Printf,
		Collision: world.OutdoorCollision})
	srv := &gameserver.Server{Auth: a, World: w, Store: st, WebDir: *webDir, Log: log.Printf}
	hs := &http.Server{Addr: *addr, Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}

	idle := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("shutting down; saving players")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hs.Shutdown(ctx)
		w.Shutdown()
		close(idle)
	}()

	log.Printf("serving %s on http://%s  (admin login %v, Discord login %v)", srv.StaticDir(), *addr, a.HasAdmin(), a.HasDiscord())
	if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	<-idle
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// loadEnvFile sets variables from KEY=VALUE lines that the environment does not already set.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if k, v, ok := strings.Cut(l, "="); ok && os.Getenv(strings.TrimSpace(k)) == "" {
			os.Setenv(strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`))
		}
	}
}

// sessionKey is the secret that signs session cookies: from the environment,
// or one generated once and kept beside the databases.
func sessionKey(dataDir string) []byte {
	if k := os.Getenv("DAOC_SESSION_KEY"); len(k) >= 16 {
		return []byte(k)
	}
	path := filepath.Join(dataDir, "session.key")
	if b, err := os.ReadFile(path); err == nil && len(b) >= 32 {
		return []byte(strings.TrimSpace(string(b)))
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	k := hex.EncodeToString(b)
	os.MkdirAll(dataDir, 0o755)
	if err := os.WriteFile(path, []byte(k), 0o600); err != nil {
		log.Printf("could not keep the session key (%v): sessions end when the server stops", err)
	}
	return []byte(k)
}
