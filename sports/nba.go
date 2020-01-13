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
	lastCheck := params.Get("date")
	if lastCheck == "" {
		lastCheck = n.DefaultTimeString()
	}
	period, _ := strconv.Atoi(params.Get("period"))
	var playByPlay gonba.PlayByPlayV2
	if period != 0 {
		playByPlay, _ = n.client.GetPlayByPlayV2(date, gameId, period)
	} else { // Get all initial plays
		playByPlay, _ = n.client.GetPlayByPlayV2All(date, gameId)
	}
	return buildResultFromNBAPlayByPlay(&playByPlay, lastCheck, period)
}

func (n *nba) ParseScheduleState(statusCode int) ScheduleState {
	if statusCode == 2 {
		return Live
	} else if statusCode == 3 {
		return Complete
	}
	return Preview
}

func (n *nba) DefaultTimeString() string {
	return "12:00"
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

func buildResultFromNBAPlayByPlay(playByPlay *gonba.PlayByPlayV2, lastCheck string, period int) map[string]interface{} {
	result := make(map[string]interface{})
	plays := make([]map[string]interface{}, 0, len(playByPlay.Plays))
	var lastPlay gonba.Play
	for _, play := range playByPlay.Plays {
		if pastLastCheck(play.Clock, lastCheck) {
			playResult := make(map[string]interface{})
			playResult["description"] = play.Formatted.Description
			playResult["typeId"] = play.EventMsgType
			playResult["periodTime"] = play.Clock
			playResult["teamId"] = play.TeamID
			playResult["playerId"] = play.PersonID
			plays = append(plays, playResult)
		}
		lastPlay = play
	}
	result["plays"] = plays
	if len(playByPlay.Plays) == 0 {
		lastPlay = gonba.Play{
			Clock: lastCheck,
		}
	}
	status := map[string]interface{}{"period": period, "periodTimeRemaining": lastPlay.Clock}
	gameState := Live
	if lastPlay.EventMsgType == 13 { // End of Quarter Event
		status["period"] = status["period"].(int) + 1
		if status["period"].(int) > 4 {
			if lastPlay.HTeamScore != lastPlay.VTeamScore {
				gameState = Complete
			}
		}
	}
	game := make(map[string]interface{})
	game["status"] = status
	game["home"] = map[string]interface{}{"score": lastPlay.HTeamScore}
	game["away"] = map[string]interface{}{"score": lastPlay.VTeamScore}
	result["game"] = game
	result["metadata"] = map[string]interface{}{"state": gameState, "lastCheck": lastPlay.Clock}
	return result
}

func pastLastCheck(playTime string, lastCheck string) bool {
	playTimeTime, _ := time.Parse("3:04", playTime)
	lastCheckTime, _ := time.Parse("3:04", lastCheck)
	return playTimeTime.Before(lastCheckTime)
	//return true
}