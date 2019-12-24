package sports

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"
const DetailedDateLayout = "2006-01-02T15:04:05Z07:00"

type Sports struct {
	MLB mlb
	NBA *nba
	NFL nfl
	NHL *nhl
}

type ScheduleState int

const (
	Preview  ScheduleState = 0
	Live     ScheduleState = 1
	Complete ScheduleState = 2
)

type GameScheduleStatus struct {
	Id            string
	StartTime     time.Time
	ScheduleState ScheduleState // 0 for preview, 1 for live, 2 for ended
}

// params[] date
type Sport interface {
	Name() string
	Schedule(params url.Values) map[string]interface{}
	PlayByPlay(params url.Values) map[string]interface{}
	ParseScheduleState(statusCode int) ScheduleState
}

func InitializeSports() *Sports {
	sports := Sports{
		NHL: InitNHL(),
		NBA: InitNBA(),
	}

	return &sports
}

func (s *Sports) RetrieveSportsAsList() [2]Sport {
	return [2]Sport{s.NBA, s.NHL}
}

func (s *Sports) ParseSportId(sportId int) Sport {
	switch sportId {
	case 0:
		//return s.MLB
	case 1:
		return s.NBA
	case 2:
		//return s.NFL
	case 3:
		return s.NHL
	}
	return nil
}

func ParseSportString(sport string) int {
	strings.ToLower(sport)
	switch sport {
	case "mlb":
		//return s.MLB
	case "nba":
		return 1
	case "nfl":
		//return s.NFL
	case "nhl":
		return 3
	}
	return -1
}

func CheckActiveGames(sport Sport, schedule map[string]interface{}) []GameScheduleStatus {
	scheduledGames := schedule["content"].([]map[string]interface{})
	games := make([]GameScheduleStatus, 0, len(scheduledGames))
	for _, scheduledGame := range scheduledGames {
		status := sport.ParseScheduleState(scheduledGame["statusCode"].(int))
		date, _ := CreateDateFromDetailedString(scheduledGame["date"].(string))
		var gameId string
		if game, ok := scheduledGame["id"].(int); ok {
			gameId = strconv.Itoa(game)
		} else {
			gameId = scheduledGame["id"].(string)
		}
		game := GameScheduleStatus{
			gameId,
			date,
			status,
		}
		games = append(games, game)
	}
	return games
}



// CreateDetailedStringFromDate converts a time.Time object to a string representing a date with format `yyyy-mm-dd`.
func CreateDetailedStringFromDate(date time.Time) string {
	return date.Format(DetailedDateLayout)
}

func CreateDateFromDetailedString(dateString string) (time.Time, error) {
	return time.Parse(DetailedDateLayout, dateString)
}

// CreateDateFromString converts a string representing a date with format `yyyy-mm-dd` to a time.Time object.
func CreateDateFromString(dateString string) (time.Time, error) {
	return time.Parse(dateLayout, dateString)
}
