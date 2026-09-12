package isolation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/domain"
	"github.com/Hikyo-Org/hikyo/internal/parameters"
	"github.com/Hikyo-Org/hikyo/internal/service"
	"github.com/Hikyo-Org/hikyo/internal/store"
)

func TestLegacyParameterizedSnapshotPinPreservesDollarPrefix(t *testing.T) {
	forEngines(t, func(t *testing.T, db *store.DB) {
		identityFixtures(t, db)
		seedDeliveryCatalogue(t, db)
		actor := service.LocalPrincipal(identAdmin)
		scope := scopeEnv(orgA, prjA1, envA1)
		publishDeliveryValues(t, db, envA1, map[string]string{"DATABASE_URL": "$${NUM}"})
		// Reconstruct the pre-escape, unversioned contract around a real encrypted
		// snapshot. Its dollar prefix is literal; the ${NUM} opening is a reference.
		declaration := `{"rule":{"type":"string","pattern":"^\\$[0-9]+$"}}`
		contract, err := json.Marshal(parameters.Contract{Declarations: map[string]string{"NUM": "[0-9]+"}, Schemas: map[string]string{"DATABASE_URL": declaration}})
		if err != nil {
			t.Fatal(err)
		}
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
		execRaw(t, db, "UPDATE snapshots SET parameter_contract="+quote(string(contract))+" WHERE environment_id='env_a1' AND revision=2")
		execRaw(t, db, "UPDATE keys SET declaration="+quote(declaration)+" WHERE id='key_fed_url'")
		execRaw(t, db, `INSERT INTO grants (id,principal_id,capability,org_id,project_id,env_id,created_at) VALUES ('g_legacy_pin','`+string(identAdmin)+`','pin','org_a','prj_a1','env_a1',`+ts+`)`)
		workload, err := identitySvc(db).CreateServiceAccount(t.Context(), actor, prjScope(), "legacy-parameter-pin", domain.ClassWorkload)
		if err != nil {
			t.Fatal(err)
		}
		grantMachineRead(t, db, workload.Principal, envA1)
		pins := &service.Pins{DB: db, Keyring: probeKeyring(t, db)}
		if _, err := pins.Set(t.Context(), actor, scope, service.SetPinRequest{WorkloadPrincipalID: workload.Principal, Revision: 2}); err != nil {
			t.Fatalf("legacy template rejected by pin preflight: %v", err)
		}
		out, err := deliverySvc(t, db).FetchAs(t.Context(), actor, scope, "", service.FetchOptions{Parameters: map[string]string{"NUM": "123"}})
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range out.Keys {
			if key.Name == "DATABASE_URL" {
				if key.Value == nil || *key.Value != "$123" {
					t.Fatalf("legacy template changed semantics: %+v", key)
				}
				return
			}
		}
		t.Fatal("legacy config absent")
	})
}
