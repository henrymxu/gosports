package sports

import (
	"net/url"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"
const resultDateLayout = "2006-01-02T15:04:05Z07:00"

type Sports struct {
	MLB mlb
	NBA nba
	NFL nfl
	NHL nhl
}

// params[] date
type Sport interface {
	Schedule(params url.Values) map[string]interface{}
	PlayByPlay(params url.Values) map[string]interface{}
}

func InitializeSports() *Sports {
	sports := Sports{
		NHL: nhl{},
	}

	return &sports
}

func (s *Sports) ParseSportId(sportId int) Sport {
	switch sportId {
	case 0:
		//return s.MLB
	case 1:
		//return s.NBA
	case 2:
		//return s.NFL
	case 3:
		return s.NHL
	}
	return nil
}

func (s *Sports) ParseSportString(sport string) Sport {
	strings.ToLower(sport)
	switch sport {
	case "mlb":
		//return s.MLB
	case "nba":
		//return s.NBA
	case "nfl":
		//return s.NFL
	case "nhl":
		return s.NHL
	}
	return nil
}

// CreateStringFromDate converts a time.Time object to a string representing a date with format `yyyy-mm-dd`.
func CreateStringFromDate(date time.Time) string {
	return date.Format(resultDateLayout)
}

// CreateDateFromString converts a string representing a date with format `yyyy-mm-dd` to a time.Time object.
func CreateDateFromString(dateString string) (time.Time, error) {
	return time.Parse(dateLayout, dateString)
}