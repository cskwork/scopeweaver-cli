package portable

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIndependentHomesSeedSkillsWithoutSharingOrResurrectingEdits(t *testing.T) {
	bundle := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bundle, "example"), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(bundle, "example", "SKILL.md")
	if err := os.WriteFile(source, []byte("bundled"), 0600); err != nil {
		t.Fatal(err)
	}
	one, two := t.TempDir(), t.TempDir()
	for _, home := range []string{one, two} {
		if err := prepareHome(home); err != nil {
			t.Fatal(err)
		}
		if _, err := seedSkills(context.Background(), bundle, home); err != nil {
			t.Fatal(err)
		}
	}
	first := filepath.Join(one, "skills", "example", "SKILL.md")
	second := filepath.Join(two, "skills", "example", "SKILL.md")
	if err := os.WriteFile(first, []byte("my edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(second); err != nil || string(b) != "bundled" {
		t.Fatalf("second home shared edits: %q %v", b, err)
	}
	if err := os.Remove(first); err != nil {
		t.Fatal(err)
	}
	if _, err := seedSkills(context.Background(), bundle, one); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("restart resurrected deleted skill: %v", err)
	}
	if b, err := os.ReadFile(source); err != nil || string(b) != "bundled" {
		t.Fatalf("bundle mutated: %q %v", b, err)
	}
}
func TestSkillsInitializationPreservesFilesFromInterruptedAttempt(t *testing.T) {
	bundle, home := t.TempDir(), t.TempDir()
	if err := prepareHome(home); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(bundle, "example"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "example", "SKILL.md"), []byte("bundled"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "example", "new.txt"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(home, "skills", "example", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(existing), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("existing user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := seedSkills(context.Background(), bundle, home); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(existing); err != nil || string(b) != "existing user edit" {
		t.Fatalf("existing file overwritten: %q %v", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(home, "skills", "example", "new.txt")); err != nil || string(b) != "new" {
		t.Fatalf("new file not copied: %q %v", b, err)
	}
}

func TestSkillsInitializationIgnoresAbandonedPartialTemporary(t *testing.T) {
	bundle, home := t.TempDir(), t.TempDir()
	if err := prepareHome(home); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{bundle, filepath.Join(home, "skills")} {
		if err := os.MkdirAll(filepath.Join(root, "example"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	content := []byte("complete bundled skill instructions")
	if err := os.WriteFile(filepath.Join(bundle, "example", "SKILL.md"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "skills", "example", ".scopeweaver-seed-abandoned"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := seedSkills(context.Background(), bundle, home); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(filepath.Join(home, "skills", "example", "SKILL.md"))
	if err != nil || string(actual) != string(content) {
		t.Fatalf("partial seed became final: %q %v", actual, err)
	}
}
func TestAtomicSeedPublicationNeverReplacesExistingFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "temporary")
	target := filepath.Join(dir, "user-file")
	if err := os.WriteFile(source, []byte("bundled"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishSeed(source, target); !os.IsExist(err) {
		t.Fatalf("publish replaced or misclassified existing target: %v", err)
	}
	actual, err := os.ReadFile(target)
	if err != nil || string(actual) != "user edit" {
		t.Fatalf("user file overwritten: %q %v", actual, err)
	}
}
