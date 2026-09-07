// host-upgrade-helper exercises the deployment adapter against real systemd.
// It deliberately does not implement Hikyo migration or release verification.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 || os.Geteuid() == 0 || os.Getegid() == 0 {
		return fmt.Errorf("helper requires an unprivileged operation")
	}
	key := os.Getenv("HIKYO_ROOT_KEY_FILE")
	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--root-key-file=") {
			key = strings.TrimPrefix(arg, "--root-key-file=")
		}
	}
	raw, err := os.ReadFile(key)
	if err != nil || string(raw) != strings.Repeat("a", 64) {
		return fmt.Errorf("runtime credential unavailable at %s: %v", key, err)
	}
	if _, err := os.ReadFile("/etc/hikyo/upgrade-keys/operator.age"); !os.IsPermission(err) {
		return fmt.Errorf("operator custody is not isolated from runtime")
	}
	if os.Getenv("OPERATOR_PASSPHRASE") != "" {
		return fmt.Errorf("operator environment leaked")
	}
	switch os.Args[1] {
	case "server":
		return http.ListenAndServe("127.0.0.1:8081", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/readyz" {
				ready, _ := os.ReadFile("ready")
				if string(ready) != "ready" {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			}
		}))
	case "migrate", "backup":
		if key != "/proc/self/fd/3" {
			return fmt.Errorf("child credential was not passed by descriptor")
		}
		fmt.Println(`{"credential":"verified","uid":"unprivileged"}`)
		return nil
	case "escrow":
		// Operator delegation: root key on descriptor 3, the root-owned custody
		// copy on descriptor 4, both environment files merged, terminal kept.
		if key != "/proc/self/fd/3" {
			return fmt.Errorf("delegated root key was not passed by descriptor")
		}
		custody := ""
		for i, arg := range os.Args {
			if arg == "--root-key-file" && i+1 < len(os.Args) {
				custody = os.Args[i+1]
			}
		}
		copyRaw, err := os.ReadFile(custody)
		if custody != "/proc/self/fd/4" || err != nil || string(copyRaw) != "custody copy" {
			return fmt.Errorf("custody copy unavailable at %q: %v", custody, err)
		}
		if os.Getenv("HIKYO_DB") != "sqlite:hikyo.db" || os.Getenv("HIKYO_UPGRADE_BUNDLE") != "/var/lib/hikyo-upgrade/bundle-fixture" {
			return fmt.Errorf("delegated environment incomplete: %q %q", os.Getenv("HIKYO_DB"), os.Getenv("HIKYO_UPGRADE_BUNDLE"))
		}
		line, err := io.ReadAll(io.LimitReader(os.Stdin, 64))
		if err != nil || strings.TrimSpace(string(line)) != "operator input" {
			return fmt.Errorf("operator stdin was not passed through: %v", err)
		}
		fmt.Println("escrow verified through delegation")
		return nil
	default:
		return fmt.Errorf("unsupported helper operation")
	}
}
