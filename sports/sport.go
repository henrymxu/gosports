package sports

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"
const DetailedDateLayout = "2006-01-02T15:04:05Z07:00"

type Sports []Sport

type ScheduleState int

const (
	Preview      ScheduleState = 0
	Live         ScheduleState = 1
	Intermission ScheduleState = 2
	Complete     ScheduleState = 3
)

type ScheduledGame struct {
	Id            string
	StartTime     time.Time
	ScheduleState ScheduleState // 0 for preview, 1 for live, 2 for ended
	Details       map[string]interface{}
}

// params[] date
type Sport interface {
	Name() string
	Schedule(params url.Values) map[string]interface{}
	PlayByPlay(params url.Values) map[string]interface{}
	ParseScheduleState(statusCode int) ScheduleState
	DefaultTimeString() string
}

func InitializeSports() *Sports {
	sports := Sports{InitNHL(), InitNBA()}
	return &sports
}

func (s *Sports) ParseSportId(sportId int) Sport {
	return (*s)[sportId]
}

func ParseSportString(sport string) int {
	strings.ToLower(sport)
	switch sport {
	case "nhl":
		return 0
	case "nba":
		return 1
	case "nfl":
		return 2
	case "mlb":
		return 3
	}
	return -1
}

func CheckActiveGames(sport Sport, schedule map[string]interface{}) []ScheduledGame {
	var scheduledGames []map[string]interface{}
	if value, ok := schedule["content"].([]map[string]interface{}); ok {
		scheduledGames = value
	} else {
		scheduledGames = make([]map[string]interface{}, 0)
	}
	games := make([]ScheduledGame, len(scheduledGames))
	for i, scheduledGame := range scheduledGames {
		status := sport.ParseScheduleState(scheduledGame["statusCode"].(int))
		date, _ := CreateDateFromDetailedString(scheduledGame["date"].(string))
		var gameId string
		if game, ok := scheduledGame["id"].(int); ok {
			gameId = strconv.Itoa(game)
		} else {
			gameId = scheduledGame["id"].(string)
		}
		game := ScheduledGame{
			Id:            gameId,
			StartTime:     date,
			ScheduleState: status,
			Details:       scheduledGame,
		}
		games[i] = game
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
