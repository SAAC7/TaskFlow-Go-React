package models

import (
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	MaxTitleLen       = 120
	MaxDescriptionLen = 1000
)

// Task is the stored entity and also what the API returns.
// The `json` tags define the API contract; the `bson` tags define the DB layout.
type Task struct {
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string        `json:"title" bson:"title"`
	Description string        `json:"description" bson:"description"`
	Completed   bool          `json:"completed" bson:"completed"`
	CreatedAt   time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt" bson:"updatedAt"`
}

// TaskInput is the body of POST and PUT. It has NO id or timestamps, so a
// client can never forge them: the server always generates them.
type TaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Validate trims the text fields and returns {field: problem}. Empty = valid.
func (in *TaskInput) Validate() map[string]string {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)

	problems := map[string]string{}
	if msg := checkTitle(in.Title); msg != "" {
		problems["title"] = msg
	}
	if msg := checkDescription(in.Description); msg != "" {
		problems["description"] = msg
	}

	return problems
}

// TaskPatch is the body of PATCH. Pointers tell apart "not sent" (nil) from
// "sent with its zero value" (e.g. {"completed": false}).
type TaskPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

// Validate checks only the fields that were actually sent.
func (p *TaskPatch) Validate() map[string]string {
	problems := map[string]string{}

	if p.Title == nil && p.Description == nil && p.Completed == nil {
		problems["body"] = "send at least one of: Title, Description, Completed"
		return problems
	}
	if p.Title != nil {
		valTitle := strings.TrimSpace(*p.Title)
		p.Title = &valTitle
		if msg := checkTitle(valTitle); msg != "" {
			problems["title"] = msg
		}
	}
	if p.Description != nil {
		valDecrip := strings.TrimSpace(*p.Description)
		p.Description = &valDecrip
		if msg := checkDescription(valDecrip); msg != "" {
			problems["description"] = msg
		}
	}

	return problems
}

// ToPatch converts a full input into a patch that sets every field (PUT).
func (in TaskInput) ToPatch() TaskPatch {
	return TaskPatch{Title: &in.Title, Description: &in.Description, Completed: &in.Completed}
}

// ListFilter holds the optional filter and the pagination of GET /tasks.
type ListFilter struct {
	Completed *bool
	Limit     int64
	Skip      int64
}

// checkTitle returns a problem description if the title is empty or longer
// than MaxTitleLen characters, or "" if the title is valid.
func checkTitle(title string) string {
	switch {
	case title == "":
		return "is required"
	case utf8.RuneCountInString(title) > MaxTitleLen:
		return "must be at most 120 characters"
	}
	return ""
}
// checkDescription returns a problem description if the description is longer
// than MaxDescriptionLen characters, or "" if it is valid (it may be empty).
func checkDescription(desc string) string {
	if utf8.RuneCountInString(desc) > MaxDescriptionLen {
		return "must be at most 1000 characters"
	}
	return ""
}
