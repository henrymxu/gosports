package sports

import (
	"fmt"
	"github.com/henrymxu/gonba"
	"github.com/ngaut/log"
	"net/url"
	"strconv"
	"time"
)

type nba struct {
	client *gonba.Client
}

func InitNBA() *nba {
	return &nba{
		client: gonba.NewClient(),
	}
}

func (n *nba) Name() string {
	return "nba"
}

func (n *nba) Schedule(params url.Values) map[string]interface{} {
	date := time.Now()
	if params != nil {
		val := string(params.Get("date"))
		if val != "" {
			tempDate, err := CreateDateFromString(val)
			if err != nil {
				log.Error("NBA schedule date error: %v", err)
			} else {
				date = tempDate
			}
		}
	}
	schedule, _ := n.client.GetSchedule(date)
	return buildResultFromNBASchedule(&schedule)
}

func (n *nba) PlayByPlay(params url.Values) map[string]interface{} {
	gameId := params.Get("gameId")
	date := time.Now()
	lastUpdate, _ := CreateDateFromDetailedString(params.Get("date"))
	period, _ := strconv.Atoi(params.Get("period"))
	var playByPlay gonba.PlayByPlayV2
	if period != 0 {
		playByPlay, _ = n.client.GetPlayByPlayV2(date, gameId, period)
	} else { // Get all initial plays
		playByPlay, _ = n.client.GetPlayByPlayV2All(date, gameId)
	}
	return buildResultFromNBAPlayByPlay(&playByPlay, &lastUpdate, period)
}

func (n *nba) ParseScheduleState(statusCode int) ScheduleState {
	return Preview
}

func buildResultFromNBASchedule(schedule *gonba.Schedule) map[string]interface{} {
	result := make(map[string]interface{})
	games := make([]map[string]interface{}, 0, len(schedule.Games))
	for _, scheduledGame := range schedule.Games {
		games = append(games, buildGameFromGames(scheduledGame))
	}
	result["content"] = games
	return result
}

func buildGameFromGames(scheduledGame gonba.Game) map[string]interface{} {
	game := make(map[string]interface{})
	game["id"] = scheduledGame.GameId
	game["date"] = CreateDetailedStringFromDate(scheduledGame.GameDate)
	game["status"] = scheduledGame.Status.StatusString
	game["statusCode"] = scheduledGame.Status.StatusCode
	game["period"] = scheduledGame.Quarter
	game["time"] = scheduledGame.QuarterTime
	game["venue"] = scheduledGame.Venue
	game["home"] = parseNBATeam(scheduledGame.Teams.Home)
	game["away"] = parseNBATeam(scheduledGame.Teams.Away)
	return game
}

func parseNBATeam(team gonba.Team) map[string]interface{} {
	result := make(map[string]interface{})
	result["teamId"] = team.Id
	result["name"] = fmt.Sprintf("%s %s", team.City, team.Name)
	result["abbr"] = team.Abbr
	result["record"] = fmt.Sprintf("%d-%d", team.Wins, team.Losses)
	result["score"] = team.Score
	return result
}

func buildResultFromNBAPlayByPlay(playByPlay *gonba.PlayByPlayV2, lastUpdate *time.Time, period int) map[string]interface{} {
	result := make(map[string]interface{})
	plays := make([]map[string]interface{}, 0, len(playByPlay.Plays))
	for _, play := range playByPlay.Plays {
		playResult := make(map[string]interface{})
		playResult["description"] = play.Formatted.Description
		playResult["typeId"] = play.EventMsgType
		playResult["periodTime"] = play.Clock
		playResult["teamId"] = play.TeamID
		playResult["playerId"] = play.PersonID
		plays = append(plays, playResult)
	}
	result["plays"] = plays
	var lastPlay gonba.Play
	if len(plays) == 0 {
		lastPlay = gonba.Play{
			Clock: "12:00",
			HTeamScore: 0,
			VTeamScore: 0,
		}
	} else {
		lastPlay = playByPlay.Plays[len(playByPlay.Plays) - 1]
	}
	status := map[string]interface{}{"period": period, "periodTimeRemaining": lastPlay.Clock}
	result["status"] = status
	if lastPlay.EventMsgType == 13 {
		status["period"] = status["period"].(int) + 1
	}
	result["home"] = map[string]interface{}{"score": lastPlay.HTeamScore}
	result["away"] = map[string]interface{}{"score": lastPlay.VTeamScore}
	result["metadata"] = map[string]interface{}{"state": Live, "timestamp": period}
	return result
}