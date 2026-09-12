package isolation

import (
	"fmt"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/service"
	"github.com/Hikyo-Org/hikyo/internal/store"
)

func TestEnvironmentDeletionReleasesOnlyScopedGrants(t *testing.T) {
	forEngines(t, func(t *testing.T, db *store.DB) {
		for _, viaDefinitions := range []bool{false, true} {
			t.Run(fmt.Sprintf("definitions=%t", viaDefinitions), func(t *testing.T) {
				f := seedDefinitionsProject(t, db, fmt.Sprintf("scoped_grants_%t", viaDefinitions), true)
				grantID := "g_" + f.env
				execRaw(t, db, fmt.Sprintf(`INSERT INTO grants (id,principal_id,capability,org_id,project_id,env_id,created_at) VALUES ('%s','%s','read','org_a','%s','%s',%s)`, grantID, bob, f.project, f.env, ts))
				seedOrigins(t, db)
				execRaw(t, db, fmt.Sprintf(`INSERT INTO grant_origins (id,grant_id,kind,subject,created_at) VALUES ('extra_%s','%s','scim','provisioning-origin',%s)`, grantID, grantID, ts))
				otherGrants := queryInt(t, db, "SELECT COUNT(*) FROM grants WHERE id <> '"+grantID+"'")
				generation := queryInt(t, db, "SELECT session_generation FROM principals WHERE id = '"+string(bob)+"'")
				if viaDefinitions {
					svc := definitionsService(t, db)
					bundle := parseDefinitions(t, exportDefinitions(t, svc, f))
					bundle.Environments = nil
					plan := planDefinitions(t, svc, f, encodeDefinitions(t, bundle))
					if _, err := svc.Apply(t.Context(), service.LocalPrincipal(alice), f.scope(), plan.ID, service.ApplyOptions{AllowDelete: true}); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := (&service.Environments{DB: db, Keyring: probeKeyring(t, db)}).Delete(t.Context(), service.LocalPrincipal(alice), f.envScope()); err != nil {
						t.Fatal(err)
					}
				}
				if n := queryInt(t, db, "SELECT COUNT(*) FROM environments WHERE id = '"+f.env+"'"); n != 0 {
					t.Fatal("environment retained")
				}
				if n := queryInt(t, db, "SELECT COUNT(*) FROM grants WHERE id = '"+grantID+"'"); n != 0 {
					t.Fatal("environment grant retained")
				}
				if n := queryInt(t, db, "SELECT COUNT(*) FROM grant_origins WHERE grant_id = '"+grantID+"'"); n != 0 {
					t.Fatal("grant origins retained")
				}
				if n := queryInt(t, db, "SELECT COUNT(*) FROM grants"); n != otherGrants {
					t.Fatal("unrelated grants changed")
				}
				if n := queryInt(t, db, "SELECT session_generation FROM principals WHERE id = '"+string(bob)+"'"); n != generation+1 {
					t.Fatal("removed authority did not invalidate principal generation")
				}
			})
		}
	})
}
