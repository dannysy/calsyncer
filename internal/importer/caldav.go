package importer

import (
	"calsyncer/internal/config"
	"fmt"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/teambition/rrule-go"
)

type CalDAV struct {
	cl *caldav.Client
}

func NewCalDAV() *CalDAV {
	var httpClient webdav.HTTPClient = http.DefaultClient
	if config.Gist().Exists(config.CALDAV_USER) && config.Gist().Exists(config.CALDAV_PASS) {
		httpClient = webdav.HTTPClientWithBasicAuth(
			httpClient,
			config.Gist().String(config.CALDAV_USER),
			config.Gist().String(config.CALDAV_PASS),
		)
	}
	cl, err := caldav.NewClient(httpClient, config.Gist().String(config.CALDAV_URL))
	if err != nil {
		log.Fatal().Err(err).Msg("error creating caldav client")
	}
	return &CalDAV{
		cl: cl,
	}
}

func (c *CalDAV) Get() (Calendar, error) {
	cl := resty.New().
		SetBasicAuth(config.Gist().String(config.CALDAV_USER), config.Gist().String(config.CALDAV_PASS))
	resp, err := cl.R().SetDoNotParseResponse(true).Get(config.Gist().String(config.CALDAV_URL))
	if err != nil {
		return Calendar{}, errors.Wrap(err, "error getting calendar")
	}
	if resp.IsError() {
		return Calendar{}, errors.New(fmt.Sprintf("error getting calendar: %s", resp.Status()))
	}
	cal, err := ics.ParseCalendar(resp.RawBody())
	calendar := Calendar{
		IDtoEvents: make(map[string]Event),
	}
	setLocation(&calendar, cal.Timezones())

	for _, event := range cal.Events() {
		calEv := Event{
			ID:           ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyUniqueId)),
			Title:        ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertySummary)),
			Description:  ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyDescription)),
			StartDateStr: ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyDtStart)),
			EndDateStr:   ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyDtEnd)),
			Location:     ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyLocation)),
			Organizer:    ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyOrganizer)),
			Rrule:        ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentPropertyRrule)),
			RecurrenceID: ValueOrEmpty(event.ComponentBase.GetProperty(ics.ComponentProperty(ics.PropertyRecurrenceId))),
			Attendies:    JoinProperties(event.Properties, ics.ComponentPropertyAttendee),
		}
		if calEv.Rrule != "" {
			rropt, err := rrule.StrToROptionInLocation(calEv.Rrule, calendar.Location)
			if err != nil {
				log.Error().Err(err).
					Str("eventID", calEv.ID).
					Str("eventTitle", calEv.Title).
					Str("eventRrule", calEv.Rrule).
					Msgf("event has invalid rrule")
			}
			dtStart, ok := calEv.GetStartAt(calendar.Location)
			if ok {
				rropt.Dtstart = dtStart
			}
			rr, err := rrule.NewRRule(*rropt)
			if err != nil {
				log.Error().Err(err).
					Str("eventID", calEv.ID).
					Str("eventTitle", calEv.Title).
					Str("eventRrule", rropt.RRuleString()).
					Msgf("event has invalid rrule")
			}
			occurances := rr.Between(time.Now().In(calendar.Location), time.Now().In(calendar.Location).Add(7*24*time.Hour), true)
			calEv.RecurrenceTimes = make([]string, 0, len(occurances))
			for _, eventDtime := range occurances {
				calEv.RecurrenceTimes = append(calEv.RecurrenceTimes, eventDtime.Format(CalendarTimeFormat))
			}
		}
		calendar.IDtoEvents[calEv.ID] = calEv
	}
	return calendar, nil
}

func setLocation(calendar *Calendar, tzones []*ics.VTimezone) {
	if len(tzones) != 0 {
		calendar.TimeZoneID = ValueOrEmpty(tzones[0].ComponentBase.GetProperty(ics.ComponentPropertyTzid))
		loc, err := time.LoadLocation(calendar.TimeZoneID)
		if err != nil {
			log.Error().Err(err).Str("timezone", calendar.TimeZoneID).Msg("error getting location by timezone id")
			calendar.Location = time.Local
		} else {
			calendar.Location = loc
		}
	} else {
		calendar.TimeZoneID = time.Local.String()
		calendar.Location = time.Local
	}
}
