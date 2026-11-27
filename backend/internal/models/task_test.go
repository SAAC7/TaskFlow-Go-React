package models_test

import (
	"strings"
	"testing"

	"taskflow-backend/internal/models"
)

// TestTaskInputValidate is a table-driven test: for each input it checks
// whether Validate accepts it or reports a problem on the expected field.
func TestTaskInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      models.TaskInput
		wantBad string // field expected to fail ("" = valid)
	}{
		{"valid", models.TaskInput{Title: "Learn Go"}, ""},
		{"blank title", models.TaskInput{Title: "   "}, "title"},
		{"title too long", models.TaskInput{Title: strings.Repeat("a", 121)}, "title"},
		{"description too long", models.TaskInput{Title: "ok", Description: strings.Repeat("a", 1001)}, "description"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems := tc.in.Validate()
			if tc.wantBad == "" && len(problems) != 0 {
				t.Fatalf("expected valid, got %v", problems)
			}
			if tc.wantBad != "" {
				if _, ok := problems[tc.wantBad]; !ok {
					t.Fatalf("expected problem on %q, got %v", tc.wantBad, problems)
				}
			}
		})
	}
}

// TestTaskInputValidateTrims checks that Validate removes the spaces around the title.
func TestTaskInputValidateTrims(t *testing.T) {
	in := models.TaskInput{Title: "  hello  "}
	in.Validate()
	if in.Title != "hello" {
		t.Fatalf("title not trimmed: %q", in.Title)
	}
}

// TestTaskPatchValidate checks the PATCH rules: an empty patch is invalid,
// {"completed": false} is a real change, and a blank title is rejected.
func TestTaskPatchValidate(t *testing.T) {
	if p := (&models.TaskPatch{}); len(p.Validate()) == 0 {
		t.Fatal("empty patch must be invalid")
	}

	done := false
	p := models.TaskPatch{Completed: &done} // {"completed": false} is a real change
	if problems := p.Validate(); len(problems) != 0 {
		t.Fatalf("completed=false must be valid, got %v", problems)
	}

	blank := " "
	if problems := (&models.TaskPatch{Title: &blank}).Validate(); problems["title"] == "" {
		t.Fatal("blank title in patch must be invalid")
	}
}
