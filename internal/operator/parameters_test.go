package operator

import (
	"bytes"
	"testing"

	"github.com/Hikyo-Org/hikyo/internal/crypto"
	hikyov1 "github.com/Hikyo-Org/hikyo/internal/operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParametersBindStampWithoutChangingMappedValues(t *testing.T) {
	r := &HikyoSecretReconciler{}
	inst := &hikyov1.HikyoInstance{ObjectMeta: metav1.ObjectMeta{UID: "instance"}}
	cr := &hikyov1.HikyoSecret{ObjectMeta: metav1.ObjectMeta{UID: "secret"}}
	cr.Spec.Target.Name = "app-config"
	root := bytes.Repeat([]byte{1}, crypto.KeySize)
	legacy, err := r.computeStamp(inst, cr, nil, root)
	if err != nil {
		t.Fatal(err)
	}
	cr.Spec.Parameters = map[string]string{"PR_NUMBER": "123"}
	first, err := r.computeStamp(inst, cr, nil, root)
	if err != nil {
		t.Fatal(err)
	}
	cr.Spec.Parameters["PR_NUMBER"] = "124"
	second, err := r.computeStamp(inst, cr, nil, root)
	if err != nil {
		t.Fatal(err)
	}
	if legacy == first || first == second {
		t.Fatal("parameters failed to move delivery stamp with identical mapped values")
	}
	cr.Spec.Parameters = map[string]string{}
	empty, err := r.computeStamp(inst, cr, nil, root)
	if err != nil {
		t.Fatal(err)
	}
	if empty != legacy {
		t.Fatal("empty params changed legacy stamp")
	}
	if bindingDigest(bindingInput{parameters: map[string]string{"PR_NUMBER": "123"}}) == bindingDigest(bindingInput{parameters: map[string]string{"PR_NUMBER": "124"}}) {
		t.Fatal("parameter change reused local cursor binding")
	}
}
