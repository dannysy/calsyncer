package exporter

import (
	"calsyncer/internal/importer"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
)

const fname = "callsyncer-export-log.txt"

type FileExporter struct {
	fileLock *sync.RWMutex
	pool     *pool.ContextPool
	next     CalExporter
}

func NewFileExporter(ctx context.Context, next CalExporter) *FileExporter {
	fe := FileExporter{next: next, fileLock: &sync.RWMutex{}, pool: pool.New().WithContext(ctx).WithMaxGoroutines(1)}
	taskr := tasker.New(tasker.Option{})
	taskr.Task("0 */1 8-20 * * 1-5 *", func(ctx context.Context) (int, error) {
		return 0, fe.deleteOldEvents()
	})
	fe.pool.Go(func(ctx context.Context) error {
		taskr.Run()
		return nil
	})
	return &fe
}

func (e *FileExporter) Set(calendar importer.Calendar) error {
	e.fileLock.RLock()
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return errors.Wrap(err, "error opening file")
	}
	defer f.Close()

	var processed, toUpdate importer.Calendar
	toUpdate.IDtoEvents = make(map[string]importer.Event)
	toUpdate.TimeZoneID = calendar.TimeZoneID
	toUpdate.Location = calendar.Location
	dec := json.NewDecoder(f)
	err = dec.Decode(&processed)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.Wrap(err, "error decoding file")
	}
	if processed.IDtoEvents == nil {
		processed.IDtoEvents = make(map[string]importer.Event)
		processed.TimeZoneID = calendar.TimeZoneID
	}
	e.fileLock.RUnlock()
	for _, event := range calendar.IDtoEvents {
		log.Debug().Msgf("processing event: %s %s", event.ID, event.Title)
		if len(event.RecurrenceTimes) == 0 {
			stAt, ok := event.GetStartAt(calendar.Location)
			stAt = stAt.UTC()
			if !ok {
				log.Warn().Msgf("event %s %s has no start date", event.ID, event.Title)
				continue
			}
			if stAt.Before(time.Now().UTC()) {
				log.Debug().Msgf("event %s %s is in the past", event.ID, event.Title)
				continue
			}
			if stAt.After(time.Now().Add(7 * 24 * time.Hour).UTC()) {
				log.Debug().Msgf("event %s %s is in the future", event.ID, event.Title)
				continue
			}
		}
		processedEvent, ok := processed.IDtoEvents[event.ID]
		if !ok {
			toUpdate.IDtoEvents[event.ID] = event
			processed.IDtoEvents[event.ID] = event
			continue
		}
		procLm, procOk := processedEvent.GetLastModified(calendar.Location)
		evLm, evOk := event.GetLastModified(calendar.Location)
		// almost always last modified time is empty
		// if (procOk != evOk) ||
		if (!procOk && !evOk && !importer.HashEqual(event.Hash(), processedEvent.Hash())) ||
			(procOk && evOk && !procLm.Equal(evLm)) {
			processed.IDtoEvents[event.ID] = event
			toUpdate.IDtoEvents[event.ID] = event
		}
	}
	f.Close()

	e.fileLock.Lock()
	defer e.fileLock.Unlock()
	rwf, err := os.OpenFile(fname, os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		return errors.Wrap(err, "error truncating file")
	}
	defer rwf.Close()

	enc := json.NewEncoder(rwf)
	enc.SetIndent("", "  ")
	err = enc.Encode(processed)
	if err != nil {
		return errors.Wrap(err, "error encoding file")
	}
	return e.next.Set(toUpdate)
}

func (e *FileExporter) deleteOldEvents() error {
	log.Debug().Str("task", "file-cleaning").Msg("tick for deleting old events")
	e.fileLock.Lock()
	defer e.fileLock.Unlock()
	f, err := os.OpenFile(fname, os.O_RDWR, 0777)
	if err != nil {
		log.Err(err).Str("task", "file-cleaning").Msg("cant open file for clean task")
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var processed importer.Calendar
	err = dec.Decode(&processed)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Err(err).Str("task", "file-cleaning").Msg("error decoding file")
		return err
	}
	processed.Location, err = time.LoadLocation(processed.TimeZoneID)
	if err != nil {
		log.Err(err).Str("task", "file-cleaning").Msg("error getting location by timezone id")
		processed.Location = time.Local
	}
	var remainingEvents = make(map[string]importer.Event)
	hasDeletions := false
	for _, event := range processed.IDtoEvents {
		remainingRecurrenceDates := make([]string, 0, len(event.RecurrenceTimes))
		for _, rtStr := range event.RecurrenceTimes {
			rt, ok := importer.GetCalendarTime(rtStr, processed.Location)
			if ok && rt.UTC().After(time.Now().UTC()) {
				remainingRecurrenceDates = append(remainingRecurrenceDates, rtStr)
			}
		}
		event.RecurrenceTimes = remainingRecurrenceDates
		stAt, ok := event.GetStartAt(processed.Location)
		if !ok {
			log.Warn().Str("task", "file-cleaning").Msgf("event %s %s has no start date", event.ID, event.Title)
			remainingEvents[event.ID] = event
			continue
		}
		if stAt.UTC().After(time.Now().UTC()) && len(event.RecurrenceTimes) > 0 {
			remainingEvents[event.ID] = event
			continue
		}
		hasDeletions = true
	}
	if !hasDeletions {
		return nil
	}

	rwf, err := os.OpenFile(fname, os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		log.Err(err).Str("task", "file-cleaning").Msg("error truncating file")
		return err
	}
	defer rwf.Close()
	enc := json.NewEncoder(rwf)
	enc.SetIndent("", "  ")
	log.Info().Str("task", "file-cleaning").Msgf("deleting %d events", len(processed.IDtoEvents)-len(remainingEvents))
	processed.IDtoEvents = remainingEvents
	err = enc.Encode(processed)
	if err != nil {
		log.Err(err).Str("task", "file-cleaning").Msg("error encoding file")
		return err
	}
	return nil
}
