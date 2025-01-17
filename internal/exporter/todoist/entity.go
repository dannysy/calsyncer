package todoist

import "time"

type Task struct {
	ID           string       `json:"id,omitempty"`
	AssignerID   any          `json:"assigner_id,omitempty"`
	AssigneeID   any          `json:"assignee_id,omitempty"`
	ProjectID    string       `json:"project_id,omitempty"`
	SectionID    any          `json:"section_id,omitempty"`
	ParentID     any          `json:"parent_id,omitempty"`
	Order        int          `json:"order,omitempty"`
	Content      string       `json:"content,omitempty"`
	Description  string       `json:"description,omitempty"`
	IsCompleted  bool         `json:"is_completed,omitempty"`
	Labels       []string     `json:"labels,omitempty"`
	Priority     int          `json:"priority,omitempty"`
	CommentCount int          `json:"comment_count,omitempty"`
	CreatorID    string       `json:"creator_id,omitempty"`
	CreatedAt    time.Time    `json:"created_at,omitempty"`
	Due          TaskDue      `json:"due,omitempty"`
	URL          string       `json:"url,omitempty"`
	Duration     TaskDuration `json:"duration,omitempty"`
	Deadline     any          `json:"deadline,omitempty"`
}

type TaskDue struct {
	Date        string `json:"date,omitempty"`
	String      string `json:"string,omitempty"`
	Lang        string `json:"lang,omitempty"`
	IsRecurring bool   `json:"is_recurring,omitempty"`
	Datetime    string `json:"datetime,omitempty"`
}

type TaskDuration struct {
	Amount int    `json:"amount,omitempty"`
	Unit   string `json:"unit,omitempty"`
}
