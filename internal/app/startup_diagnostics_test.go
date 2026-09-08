package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/diagnostics"
)

func TestServerDiagnosticsLevelsAndReadiness(t *testing.T) {
	for level := 0; level <= 3; level++ {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			cfg := devConfig(t)
			cfg.Store.DSN = "postgres://operator:DSN_SECRET_SENTINEL@example.invalid/database"
			var output bytes.Buffer
			ctx := diagnostics.With(t.Context(), level, &output)
			srv, err := Boot(ctx, cfg, testLogger())
			if err != nil {
				t.Fatal(err)
			}
			// Boot has acquired sockets but ServeWithReady has not started them.
			if strings.Contains(output.String(), "Server is accepting requests") {
				_ = srv.Close()
				t.Fatal("boot claimed serving readiness")
			}
			serveCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			if err := srv.ServeWithReady(serveCtx, cancel); err != nil {
				t.Fatal(err)
			}
			logged := output.String()
			if level == 0 {
				if logged != "" {
					t.Fatalf("default diagnostics: %q", logged)
				}
				return
			}
			for _, want := range []string{"Loading root key", "Opening admitted datastore", "Loading encryption keyring", "Resolving managed deployment configuration", "Preparing public and operational listeners", "Preparing runtime services and providers", "Server boot finished; listeners prepared", "Server is accepting requests"} {
				if !strings.Contains(logged, want) {
					t.Errorf("missing phase %q: %s", want, logged)
				}
			}
			if strings.Contains(logged, "Datastore pools:") != (level >= 2) {
				t.Errorf("wrong detail level: %s", logged)
			}
			if strings.Contains(logged, "server boot elapsed=") != (level >= 3) {
				t.Errorf("wrong timing level: %s", logged)
			}
			root, err := os.ReadFile(devRootKeyPath(cfg))
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"DSN_SECRET_SENTINEL", cfg.Store.DSN, cfg.Store.Path, strings.TrimSpace(string(root))} {
				if secret != "" && strings.Contains(logged, secret) {
					t.Fatal("diagnostics exposed configuration or key material")
				}
			}
		})
	}
}

func TestServerDiagnosticsDoNotClaimSuccessAfterListenerFailure(t *testing.T) {
	var output bytes.Buffer
	ctx := diagnostics.With(t.Context(), 3, &output)
	resources := recordingBootResources(&bootResourceRecord{})
	injected := errors.New("LISTENER_SECRET_SENTINEL")
	resources.listen = func(string, string) (net.Listener, error) { return nil, injected }
	_, err := boot(ctx, devConfig(t), testLogger(), resources)
	if !errors.Is(err, injected) {
		t.Fatalf("changed startup error: %v", err)
	}
	logged := output.String()
	if !strings.Contains(logged, "Preparing public and operational listeners") {
		t.Fatal("missing attempted phase")
	}
	for _, absent := range []string{"LISTENER_SECRET_SENTINEL", "Public and operational sockets are bound", "Preparing runtime services and providers", "Server boot finished", "Server is accepting requests"} {
		if strings.Contains(logged, absent) {
			t.Fatalf("unexpected diagnostic %q", absent)
		}
	}
}

func TestMigrateDiagnosticsLevelsAndNoOp(t *testing.T) {
	for level := 0; level <= 3; level++ {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			cfg := devConfig(t)
			cfg.Store.DSN = "postgres://operator:MIGRATE_SECRET_SENTINEL@example.invalid/db"
			var output bytes.Buffer
			ctx := diagnostics.With(t.Context(), level, &output)
			if err := RunMigrate(ctx, cfg, testLogger()); err != nil {
				t.Fatal(err)
			}
			logged := output.String()
			if level == 0 && logged != "" {
				t.Fatalf("default diagnostics: %q", logged)
			}
			if level > 0 {
				for _, want := range []string{"Preparing explicit schema migration", "Verifying signed release bundle", "Applying schema migrations", "Verifying migrated schema", "Schema migration finished"} {
					if !strings.Contains(logged, want) {
						t.Errorf("missing phase %q: %s", want, logged)
					}
				}
			}
			if strings.Contains(logged, "engine=sqlite") != (level >= 2) {
				t.Errorf("wrong detail level: %s", logged)
			}
			if strings.Contains(logged, "apply schema migrations elapsed=") != (level >= 3) {
				t.Errorf("wrong timing level: %s", logged)
			}
			if strings.Contains(logged, "MIGRATE_SECRET_SENTINEL") || strings.Contains(logged, cfg.Store.Path) {
				t.Fatal("configuration exposed")
			}
			output.Reset()
			if err := RunMigrate(ctx, cfg, testLogger()); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(output.String(), "Applying schema migrations") {
				t.Fatal("no-op migrate claimed to apply SQL")
			}
		})
	}
}

func TestMigrateDiagnosticsRefusalDoesNotClaimWrites(t *testing.T) {
	cfg := devConfig(t)
	cfg.Dev = false
	var output bytes.Buffer
	err := RunMigrate(diagnostics.With(t.Context(), 3, &output), cfg, testLogger())
	if err == nil {
		t.Fatal("untrusted migration admitted")
	}
	for _, absent := range []string{"Applying schema migrations", "Schema migration finished"} {
		if strings.Contains(output.String(), absent) {
			t.Fatalf("refused migration claimed %q", absent)
		}
	}
	if _, err := os.Stat(cfg.Store.Path); !os.IsNotExist(err) {
		t.Fatal("refused migration created database", err)
	}
}
