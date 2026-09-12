package isolation

import (
	"errors"
	"testing"
	"time"

	"github.com/Hikyo-Org/hikyo/internal/service"
	"github.com/Hikyo-Org/hikyo/internal/store"
)

func TestExportAuditSeparatesPublicParametersFromSecretDisclosure(t *testing.T) {
	forEngines(t, func(t *testing.T, db *store.DB) {
		identityFixtures(t, db)
		seedDeliveryCatalogue(t, db)
		actor := service.LocalPrincipal(identAdmin)
		scope := scopeEnv(orgA, prjA1, envA1)
		svc := revisionSvc(t, db)
		configBefore := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.value_revealed' AND object_id='key_fed_url'`)
		secretBefore := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.value_revealed' AND object_id='key_fed_pw'`)
		exportBefore := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported'`)
		if _, _, err := svc.Export(t.Context(), actor, scope, 0, false); err != nil {
			t.Fatal(err)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported'`); n != exportBefore {
			t.Fatal("ordinary config export emitted parameter audit")
		}

		execRaw(t, db, `INSERT INTO grants (id,principal_id,capability,org_id,project_id,env_id,created_at) VALUES ('g_export_reveal','`+string(identAdmin)+`','reveal','org_a','prj_a1',NULL,`+ts+`)`)
		envs := &service.Environments{DB: db, Keyring: probeKeyring(t, db)}
		if err := envs.SetParameter(t.Context(), actor, scope, "PR_NUMBER", "^[0-9]+$", false); err != nil {
			t.Fatal(err)
		}
		publishDeliveryValues(t, db, envA1, map[string]string{"DATABASE_URL": "https://pr-${PR_NUMBER}.example.com"})
		supplied := map[string]string{"PR_NUMBER": "123"}
		for _, reveal := range []bool{false, true} {
			if _, _, err := svc.ExportWithParameters(t.Context(), actor, scope, 0, reveal, supplied); err != nil {
				t.Fatal(err)
			}
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.value_revealed' AND object_id='key_fed_url'`); n != configBefore {
			t.Fatalf("config export added per-key disclosure events: %d -> %d", configBefore, n)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.value_revealed' AND object_id='key_fed_pw'`); n != secretBefore+1 {
			t.Fatalf("secret export disclosure count = %d, want %d", n, secretBefore+1)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported' AND payload LIKE '%"PR_NUMBER":"123"%' AND payload LIKE '%"revision":2%'`); n != 2 {
			t.Fatalf("parameter export audit count = %d, want 2", n)
		}
		if _, _, err := svc.ExportWithParameters(t.Context(), actor, scope, 0, false, map[string]string{"PR_NUMBER": "invalid"}); err == nil {
			t.Fatal("invalid inputs exported")
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported'`); n != exportBefore+2 {
			t.Fatal("refused export emitted success event")
		}
	})
}

func TestParameterizedExportConsentDoesNotDuplicateConfigDisclosure(t *testing.T) {
	forEngines(t, func(t *testing.T, db *store.DB) {
		fixture := ceremonyFixture(t, db, "parameter-export-ceremony")
		auth := fixture.admin.auth
		auth.ReauthWindow = 5 * time.Minute
		auth.ReauthHardCap = time.Hour
		scope := scopeEnv(orgA, prjA1, envA1)
		envs := &service.Environments{DB: db, Keyring: fixture.values.Keyring}
		if err := envs.SetParameter(t.Context(), service.LocalPrincipal(custodian), scope, "PR_NUMBER", "^[0-9]+$", false); err != nil {
			t.Fatal(err)
		}
		execRaw(t, db, `INSERT INTO keys (id,org_id,project_id,name,folder_path,classification,description,deprecated,deprecation_note,declaration,required_mode,forbidden_mode,group_id,created_at) VALUES ('key_export_config','org_a','prj_a1','EXPORT_CONFIG','','config','',FALSE,'','{"rule":{"type":"string"}}','none','none',NULL,`+ts+`)`)
		publishValue(t, fixture.values, service.LocalPrincipal(custodian), scope, "EXPORT_CONFIG", "preview-${PR_NUMBER}")
		revisions := &service.Revisions{DB: db, Keyring: fixture.values.Keyring, Auth: auth}
		supplied := map[string]string{"PR_NUMBER": "321"}
		before := disclosureRows(t, db)
		// CLI's consent preparation fetch exposes config, then the real reveal
		// must pass a fresh ceremony. Config must never enter per-key disclosure.
		if _, _, err := revisions.ExportWithParameters(t.Context(), service.Bearer(fixture.admin.token), scope, 0, false, supplied); err != nil {
			t.Fatal(err)
		}
		if _, _, err := revisions.ExportWithParameters(t.Context(), service.Bearer(fixture.admin.token), scope, 0, true, supplied); !errors.Is(err, service.ErrNoReauthWindow) {
			t.Fatalf("reveal without ceremony = %v", err)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported'`); n != 1 {
			t.Fatalf("consent probe/blocked reveal export events = %d, want 1", n)
		}
		res := passkeyCeremony(t, auth, t.Context(), fixture.admin.token, service.PurposeReveal, string(envA1), nil, fixture.device)
		if _, _, err := revisions.ExportWithParameters(t.Context(), service.Bearer(res.SessionToken), scope, 0, true, supplied); err != nil {
			t.Fatal(err)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.value_revealed' AND object_id='key_export_config'`); n != 0 {
			t.Fatalf("config disclosure events = %d, want 0", n)
		}
		if got := disclosureRows(t, db); got != before+2 {
			t.Fatalf("secret disclosure events added = %d, want 2", got-before)
		}
		if n := queryInt(t, db, `SELECT COUNT(*) FROM audit_tenant_events WHERE type='disclosure.values_exported'`); n != 2 {
			t.Fatalf("successful export events = %d, want 2", n)
		}
	})
}
