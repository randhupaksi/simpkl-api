package database

import (
	"reflect"
	"testing"
)

func TestEmbeddedMigrationsAreOrdered(t *testing.T) {
	items, err := embeddedMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 3 {
		t.Fatalf("expected the current migration set, got %d files", len(items))
	}
	versions := make([]int, len(items))
	for index, item := range items {
		versions[index] = item.Version
	}
	if !reflect.DeepEqual(versions, []int{1, 2, 3}) {
		t.Fatalf("unexpected migration versions: %#v", versions)
	}
}

func TestSplitSQLStatementsKeepsSemicolonsInsideStrings(t *testing.T) {
	statements := splitSQLStatements(`INSERT INTO example (label) VALUES ('A;B');
CREATE TABLE example_two (id INT);`)
	if len(statements) != 2 {
		t.Fatalf("expected two statements, got %d: %#v", len(statements), statements)
	}
	if statements[0] != "INSERT INTO example (label) VALUES ('A;B')" {
		t.Fatalf("string literal was split incorrectly: %s", statements[0])
	}
}
