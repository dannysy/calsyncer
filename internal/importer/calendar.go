package importer

import (
	"crypto/md5"
	"reflect"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/rs/zerolog/log"
)

const CalendarTimeFormat = "20060102T150405"

type Calendar struct {
	IDtoEvents map[string]Event `json:"events"`
	TimeZoneID string           `json:"timezoneid"`
	Location   *time.Location   `json:"-"`
}

type Event struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	LastModifiedStr string   `json:"lastModified"`
	StartDateStr    string   `json:"startDate"`
	EndDateStr      string   `json:"endDate"`
	Location        string   `json:"location"`
	Organizer       string   `json:"organizer"`
	Rrule           string   `json:"rrule"`
	RecurrenceTimes []string `json:"recurrenceTimes"`
	RecurrenceID    string   `json:"recurrenceID"`
	Attendies       string   `json:"attendies"`
}

func (e Event) GetStartAt(location *time.Location) (startAt time.Time, ok bool) {
	return GetCalendarTime(e.StartDateStr, location)
}

func (e Event) GetEndAt(location *time.Location) (endAt time.Time, ok bool) {
	return GetCalendarTime(e.EndDateStr, location)
}

func (e Event) GetLastModified(location *time.Location) (lastModified time.Time, ok bool) {
	return GetCalendarTime(e.LastModifiedStr, location)
}

func (e Event) Hash() [16]byte {
	rv := reflect.ValueOf(e)
	var stuckedFields string
	for i := 0; i < rv.NumField(); i++ {
		switch rv.Field(i).Kind() {
		case reflect.String:
			stuckedFields += rv.Field(i).String()
		case reflect.Slice:
			sl, ok := rv.Field(i).Interface().([]string)
			if !ok {
				log.Fatal().Msgf("Event struct unknown field type: %s", rv.Field(i).Kind())
			}
			stuckedFields += strings.Join(sl, ",")
		default:
			log.Fatal().Msgf("Event struct unknown field type: %s", rv.Field(i).Kind())
		}
	}
	buf := []byte(stuckedFields)
	return md5.Sum(buf)
}

func GetCalendarTime(timeStr string, location *time.Location) (out time.Time, ok bool) {
	out, err := time.ParseInLocation(CalendarTimeFormat, timeStr, location)
	if err != nil {
		return time.Time{}, false
	}
	return out, true
}

func ValueOrEmpty(prop *ics.IANAProperty) string {
	if prop == nil {
		return ""
	}
	return prop.Value
}

func HashEqual(a, b [16]byte) bool {
	for i := 0; i < 16; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func JoinProperties(props []ics.IANAProperty, propName ics.ComponentProperty) string {
	sb := strings.Builder{}
	for _, prop := range props {
		if prop.IANAToken == string(propName) {
			_, _ = sb.WriteString(prop.Value)
			_, _ = sb.WriteRune(',')
		}
	}
	if sb.Len() == 0 {
		return ""
	}
	return sb.String()[:sb.Len()-1]
}
