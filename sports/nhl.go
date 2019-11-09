package sports

import (
	"fmt"
	"github.com/henrymxu/gonhl"
	"github.com/ngaut/log"
	"net/url"
	"strconv"
)

type nhl struct {
	client *gonhl.Client
}

func (n nhl) Name() string {
	return "nhl"
}

func (n nhl) Schedule(params url.Values) map[string]interface{} {
	if n.client == nil {
		n.client = gonhl.NewClient()
	}
	schedule, _ := n.client.GetSchedule(buildScheduleParamsFromParams(params))
	return buildResultFromSchedule(&schedule)
}

func (n nhl) PlayByPlay(params url.Values) map[string]interface{} {
	if n.client == nil {
		n.client = gonhl.NewClient()
	}
	id, _ := strconv.Atoi(params.Get("gameId"))
	time, _ := CreateDateFromString(params.Get("date"))
	liveData, _ := n.client.GetGameLiveDataDiff(id, time)
	log.Debugf("Retrieved LiveData from server")
	return buildResultFromLiveData(&liveData)
}

func (n nhl) CheckActiveGames(schedule map[string]interface{}) []Game {
	scheduledGames := schedule["content"].([]map[string]interface{})
	games := make([]Game, 0, len(scheduledGames))
	for _, scheduledGame := range scheduledGames {
		status := 0
		statusCode := scheduledGame["statusCode"].(int)
		if statusCode >= 5 { // 5 6 7 states
			status = 2
		} else if statusCode > 2 { // 3 4 states
			status = 1
		}
		game := Game {
			scheduledGame["id"].(int),
			status,
		}
		games = append(games, game)
	}
	return games
}

func buildScheduleParamsFromParams(params url.Values) *gonhl.ScheduleParams {
	scheduleParams := gonhl.NewScheduleParams()
	scheduleParams.ShowLinescore()
	if params != nil {
		val := string(params.Get("date"))
		if val != "" {
			date, err := CreateDateFromString(val)
			if err == nil {
				scheduleParams.SetDate(date)
			}
		}
	}
	return scheduleParams
}

func buildResultFromLiveData(liveData *gonhl.LiveData) map[string]interface{} {
	result := make(map[string]interface{})
	result["game"] = buildGameFromLiveData(liveData)
	result["plays"] = buildPlaysFromPlays(&liveData.Plays)
	result["players"] = buildPlayersFromBoxScore(&liveData.Boxscore)
	return result
}

func buildGameFromLiveData(liveData *gonhl.LiveData) map[string]interface{} {
	game := make(map[string]interface{})
	game["home"] = buildTeamFromLinescore(&liveData.Linescore.Teams.Home)
	game["away"] = buildTeamFromLinescore(&liveData.Linescore.Teams.Away)
	game["status"] = buildStatusFromLinescore(&liveData.Linescore)

	return game
}

func buildStatusFromLinescore(linescore *gonhl.Linescore) map[string]interface{} {
	status := make(map[string]interface{})
	status["period"] = linescore.CurrentPeriod
	status["periodTimeRemaining"] = linescore.CurrentPeriodTimeRemaining

	return status
}

func buildTeamFromLinescore(linescoreTeam *gonhl.LinescoreTeam) map[string]interface{} {
	team := make(map[string]interface{})
	team["name"] = linescoreTeam.Team.Name
	team["score"] = linescoreTeam.Goals
	team["shots"] = linescoreTeam.ShotsOnGoal

	return team
}

func buildPlaysFromPlays(playsData *gonhl.Plays) []map[string]interface{} {
	plays := make([]map[string]interface{}, 0, len(playsData.AllPlays))
	for _, playData := range playsData.AllPlays {
		play := make(map[string]interface{})
		play["description"] = playData.Result.Description
		play["typeId"] = playData.Result.EventTypeID
		play["periodTime"] = playData.About.PeriodTime
		play["coordinates"] = map[string]int{"x": playData.Coordinates.X, "y": playData.Coordinates.Y}
		plays = append(plays, play)
	}
	return plays
}

func buildPlayersFromBoxScore(boxscore *gonhl.Boxscore) map[string]interface{} {
	players := make(map[string]interface{}, 0)
	players["home"] = buildPlayersFromTeam(boxscore.Teams.Home)
	players["away"] = buildPlayersFromTeam(boxscore.Teams.Away)
	return players
}

func buildPlayersFromTeam(team gonhl.BoxscoreTeam) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(team.OnIcePlus))
	for _, onIceSkater := range team.OnIcePlus {
		skater := make(map[string]interface{})
		stringSkaterId := fmt.Sprintf("ID%d", onIceSkater.PlayerID)
		skaterData := team.Players[stringSkaterId]
		skater["id"] = skaterData.Person.ID
		skater["onIceDuration"] = onIceSkater.ShiftDuration
		skater["name"] = skaterData.Person.FullName
		skater["number"] = skaterData.JerseyNumber
		skater["position"] = skaterData.Position.Abbreviation

		result = append(result, skater)
	}
	return result
}

func buildResultFromSchedule(schedule *gonhl.Schedule) map[string]interface{} {
	result := make(map[string]interface{})
	result["content"] = []string{}
	if len(schedule.Dates) == 0 {
		return result
	}
	resultGames := make([]map[string]interface{}, 0, schedule.Dates[0].TotalGames)
	games := schedule.Dates[0].Games
	for _, game := range games {
		resultGame := make(map[string]interface{})
		resultGame["id"] = game.GamePk
		resultGame["date"] = CreateStringFromDate(game.GameDate)
		resultGame["status"] = game.Status.AbstractGameState
		resultGame["statusCode"] = game.Status.CodedGameState
		resultGame["period"] = game.Linescore.CurrentPeriod
		resultGame["time"] = game.Linescore.CurrentPeriodTimeRemaining
		resultGame["home"] = parseTeam(game.Teams.Home)
		resultGame["away"] = parseTeam(game.Teams.Away)
		resultGame["venue"] = game.Venue.Name
		resultGames = append(resultGames, resultGame)
	}
	result["content"] = resultGames
	return result
}

func parseTeam(team gonhl.GameTeam) map[string]interface{} {
	resultTeam := make(map[string]interface{})
	resultTeam["teamId"] = team.Team.ID
	resultTeam["name"] = team.Team.Name
	resultTeam["abbr"] = team.Team.ShortName
	resultTeam["record"] = fmt.Sprintf("%d-%d-%d", team.LeagueRecord.Wins, team.LeagueRecord.Losses, team.LeagueRecord.Ot)
	resultTeam["score"] = team.Score
	return resultTeam
}

