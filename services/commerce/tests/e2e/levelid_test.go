package e2e

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestLevelID_Collision(t *testing.T) {
	// simulate Gen 2 logic

	org1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenant1 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	org2 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenant2 := uuid.MustParse("00000000-0000-0000-0000-000000000003") // different tenant

	loc1 := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("default-location-%s-%s", org1.String(), tenant1.String())))
	loc2 := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("default-location-%s-%s", org2.String(), tenant2.String())))

	variantID := uuid.NewMD5(uuid.NameSpaceOID, []byte("product-1"))

	level1 := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%s-%s-%s", org1.String(), tenant1.String(), loc1.String(), variantID.String())))
	level2 := uuid.NewMD5(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%s-%s-%s", org2.String(), tenant2.String(), loc2.String(), variantID.String())))

	if level1 == level2 {
		t.Errorf("Cross-tenant collision detected! %v == %v", level1, level2)
	} else {
		t.Logf("No collision: \nTenant 1 level: %v\nTenant 2 level: %v", level1, level2)
	}
}
