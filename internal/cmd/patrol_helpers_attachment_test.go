package cmd

import (
	"testing"

	"github.com/steveyegge/gastown/internal/beads"
)

func TestFindActivePatrol_AttachedFormulaMatchesWhenTitleDoesNot(t *testing.T) {
	requireBd(t)
	tmpDir, b := setupPatrolTestDB(t)

	assignee := "testrig/refinery"
	root, err := b.Create(beads.CreateOptions{
		Title:    "standalone patrol cycle",
		Priority: -1,
	})
	if err != nil {
		t.Fatalf("create patrol root: %v", err)
	}

	hooked := beads.StatusHooked
	desc := beads.FormatAttachmentFields(&beads.AttachmentFields{
		AttachedFormula: "mol-refinery-patrol",
		AttachedVars:    []string{"target_branch=main"},
		FormulaVars:     "target_branch=main",
	})
	if err := b.Update(root.ID, beads.UpdateOptions{
		Status:      &hooked,
		Assignee:    &assignee,
		Description: &desc,
	}); err != nil {
		t.Fatalf("hook patrol: %v", err)
	}

	cfg := PatrolConfig{
		PatrolMolName: "mol-refinery-patrol",
		BeadsDir:      tmpDir,
		Assignee:      assignee,
		Beads:         b,
	}

	patrolID, _, found, findErr := findActivePatrol(cfg)
	if findErr != nil {
		t.Fatalf("findActivePatrol error: %v", findErr)
	}
	if !found {
		t.Fatal("expected to find active patrol via attached_formula")
	}
	if patrolID != root.ID {
		t.Errorf("patrolID = %q, want %q", patrolID, root.ID)
	}
}

func TestBurnPreviousPatrolWisps_AttachedFormulaMatchesWhenTitleDoesNot(t *testing.T) {
	requireBd(t)
	tmpDir, b := setupPatrolTestDB(t)

	assignee := "testrig/refinery"
	root, err := b.Create(beads.CreateOptions{
		Title:    "standalone patrol cycle",
		Priority: -1,
	})
	if err != nil {
		t.Fatalf("create patrol root: %v", err)
	}

	hooked := beads.StatusHooked
	desc := beads.FormatAttachmentFields(&beads.AttachmentFields{
		AttachedFormula: "mol-refinery-patrol",
	})
	if err := b.Update(root.ID, beads.UpdateOptions{
		Status:      &hooked,
		Assignee:    &assignee,
		Description: &desc,
	}); err != nil {
		t.Fatalf("hook patrol: %v", err)
	}

	cfg := PatrolConfig{
		PatrolMolName: "mol-refinery-patrol",
		BeadsDir:      tmpDir,
		Assignee:      assignee,
		Beads:         b,
	}

	burnPreviousPatrolWisps(cfg)

	issue, err := b.Show(root.ID)
	if err != nil {
		t.Fatalf("show patrol: %v", err)
	}
	if issue.Status != "closed" {
		t.Errorf("patrol status = %q, want closed", issue.Status)
	}
}
