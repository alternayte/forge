package migrate

import (
	"strings"
	"testing"
)

func TestContainsDestructiveChange_DropTable(t *testing.T) {
	sql := `
CREATE TABLE users (id INT);
DROP TABLE old_users;
`
	if !ContainsDestructiveChange(sql) {
		t.Error("Expected DROP TABLE to be detected as destructive")
	}
}

func TestContainsDestructiveChange_DropColumn(t *testing.T) {
	sql := `
ALTER TABLE users DROP COLUMN email;
`
	if !ContainsDestructiveChange(sql) {
		t.Error("Expected DROP COLUMN to be detected as destructive")
	}
}

func TestContainsDestructiveChange_AlterType(t *testing.T) {
	sql := `
ALTER COLUMN status TYPE varchar(50);
`
	if !ContainsDestructiveChange(sql) {
		t.Error("Expected ALTER COLUMN TYPE to be detected as destructive")
	}
}

func TestContainsDestructiveChange_DropIndex(t *testing.T) {
	sql := `
DROP INDEX idx_users_email;
`
	if !ContainsDestructiveChange(sql) {
		t.Error("Expected DROP INDEX to be detected as destructive")
	}
}

func TestContainsDestructiveChange_SafeChanges(t *testing.T) {
	sql := `
CREATE TABLE users (
  id uuid PRIMARY KEY,
  name varchar(255)
);

ALTER TABLE users ADD COLUMN email varchar(255);

CREATE INDEX idx_users_email ON users(email);
`
	if ContainsDestructiveChange(sql) {
		t.Error("Expected safe changes to NOT be flagged as destructive")
	}
}

func TestFindDestructiveChanges_MultipleMatches(t *testing.T) {
	sql := `
CREATE TABLE products (id INT);
DROP COLUMN name;
ALTER TABLE orders DROP COLUMN status;
DROP TABLE old_products;
DROP INDEX idx_old;
`
	changes := FindDestructiveChanges(sql)

	// Should find 4 destructive operations
	if len(changes) != 4 {
		t.Errorf("Expected 4 destructive changes, got %d", len(changes))
	}

	// Verify each change includes line number
	for _, change := range changes {
		if !strings.Contains(change, "(line ") {
			t.Errorf("Expected change to include line number, got: %s", change)
		}
	}

	// Verify specific patterns are found
	foundDropColumn := false
	foundDropTable := false
	foundDropIndex := false

	for _, change := range changes {
		if strings.Contains(strings.ToUpper(change), "DROP COLUMN") {
			foundDropColumn = true
		}
		if strings.Contains(strings.ToUpper(change), "DROP TABLE") {
			foundDropTable = true
		}
		if strings.Contains(strings.ToUpper(change), "DROP INDEX") {
			foundDropIndex = true
		}
	}

	if !foundDropColumn {
		t.Error("Expected to find DROP COLUMN in changes")
	}
	if !foundDropTable {
		t.Error("Expected to find DROP TABLE in changes")
	}
	if !foundDropIndex {
		t.Error("Expected to find DROP INDEX in changes")
	}
}

func TestDestructiveWarning_Format(t *testing.T) {
	changes := []string{
		"DROP COLUMN name (line 5)",
		"DROP TABLE old_products (line 12)",
	}

	warning := DestructiveWarning(changes)

	// Verify warning contains expected content
	requiredStrings := []string{
		"WARNING",
		"Destructive migration",
		"permanently delete data",
		"DROP COLUMN name (line 5)",
		"DROP TABLE old_products (line 12)",
		"forge migrate diff --force",
		"review your schema changes",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(warning, required) {
			t.Errorf("Expected warning to contain '%s', but it didn't.\nWarning:\n%s", required, warning)
		}
	}

	// Verify CERTAIN is emphasized (should be in bold style)
	if !strings.Contains(warning, "CERTAIN") {
		t.Error("Expected warning to emphasize CERTAIN")
	}
}

func TestDestructiveWarning_EmptyChanges(t *testing.T) {
	changes := []string{}
	warning := DestructiveWarning(changes)

	// Should still produce a valid warning even with no changes
	if !strings.Contains(warning, "WARNING") {
		t.Error("Expected warning header even with empty changes")
	}
}

func TestContainsDestructiveChange_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name string
		sql  string
	}{
		{"lowercase", "drop table users;"},
		{"uppercase", "DROP TABLE users;"},
		{"mixed case", "DrOp TaBlE users;"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !ContainsDestructiveChange(tc.sql) {
				t.Errorf("Expected case-insensitive detection for: %s", tc.sql)
			}
		})
	}
}
