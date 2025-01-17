package exporter

import (
	"calsyncer/internal/config"
	"calsyncer/internal/exporter/todoist"
	"calsyncer/internal/importer"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var _ CalExporter = (*Todoist)(nil)

const (
	baseURL   = "https://api.todoist.com/rest/v2"
	tasksPath = "/tasks"
)

type Todoist struct {
	rc *resty.Client
}

type EventIdToTodoistTask map[string]todoist.Task

func NewTodoist() *Todoist {
	return &Todoist{
		rc: resty.New().
			SetBaseURL(baseURL).
			SetAuthToken(config.Gist().String(config.TODOIST_TOKEN)),
	}
}

func (t *Todoist) Set(c importer.Calendar) error {
	tasks, err := t.getTasks()
	if err != nil {
		return err
	}
	for _, event := range c.IDtoEvents {
		task, found := tasks[event.ID]
		if !found {
			err = t.recPostTask(&event, tasksPath, c.Location)
			if err != nil {
				log.Err(err).
					Str("eventID", event.ID).
					Str("eventTitle", event.Title).
					Msg("error creating task")
			}
			continue
		}
		err = t.updateTask(&event, task.ID, c.Location)
		if err != nil {
			log.Err(err).
				Str("eventID", event.ID).
				Str("eventTitle", event.Title).
				Msg("error creating task")
		}
	}
	return nil
}

func (t *Todoist) recPostTask(event *importer.Event, path string, location *time.Location) error {
	if len(event.RecurrenceTimes) == 0 {
		return t.postTask(event, path, location)
	}
	stAt, stAtOk := event.GetStartAt(location)
	endAt, endAtOk := event.GetEndAt(location)
	var duration time.Duration
	if stAtOk && endAtOk {
		duration = endAt.Sub(stAt)
	}
	for _, dtStartStr := range event.RecurrenceTimes {
		event.StartDateStr = dtStartStr
		if duration != 0 {
			stAt, _ := importer.GetCalendarTime(dtStartStr, location)
			event.EndDateStr = stAt.Add(duration).Format(importer.CalendarTimeFormat)
		}
		err := t.postTask(event, path, location)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Todoist) postTask(event *importer.Event, path string, location *time.Location) error {
	task := map[string]string{}
	task["content"] = event.Title
	task["project_id"] = config.Gist().String(config.TODOIST_PROJECTID)
	stAt, stAtOk := event.GetStartAt(location)
	endAt, endAtOk := event.GetEndAt(location)
	if stAtOk && endAtOk {
		durationMin := int(endAt.UTC().Sub(stAt.UTC()).Minutes())
		task["due_datetime"] = stAt.UTC().Format(time.RFC3339)
		task["duration"] = fmt.Sprintf("%d", durationMin)
		task["duration_unit"] = "minute"
	}
	eventJson, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshaling event")
	}
	task["description"] = base64.StdEncoding.EncodeToString(eventJson)
	resp, err := t.rc.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Request-Id", uuid.NewString()).
		SetBody(task).
		Post(path)
	if err != nil {
		return errors.Wrap(err, "error creating todoist task")
	}
	if resp.IsError() {
		return errors.New(fmt.Sprintf("error creating todoist task: %s", resp.Status()))
	}
	return nil
}

func (t *Todoist) updateTask(event *importer.Event, taskId string, location *time.Location) error {
	return t.recPostTask(event, tasksPath+"/"+taskId, location)
}

func (t *Todoist) getTasks() (EventIdToTodoistTask, error) {
	var tasks []todoist.Task
	resp, err := t.rc.R().
		SetQueryParam("project_id", config.Gist().String(config.TODOIST_PROJECTID)).
		SetResult(&tasks).
		Get(tasksPath)
	if err != nil {
		return nil, errors.Wrap(err, "error getting todoist tasks")
	}
	if resp.IsError() {
		return nil, errors.New(fmt.Sprintf("error getting todoist tasks: %s", resp.Status()))
	}
	out := make(EventIdToTodoistTask, len(tasks))
	for _, task := range tasks {
		type Identity struct {
			ID string `json:"id"`
		}

		var event Identity
		eventBytes, err := base64.StdEncoding.DecodeString(task.Description)
		if err != nil {
			log.Error().Msgf("task %s:%s description is not base64 encoded", task.ID, task.Content)
		}
		err = json.Unmarshal(eventBytes, &event)
		if err != nil {
			log.Error().Msgf("task %s:%s has no event in description", task.ID, task.Content)
			continue
		}
		out[event.ID] = task
	}
	return out, nil
}
