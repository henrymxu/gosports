package sports

import (
	"fmt"
	"github.com/henrymxu/gonhl"
	"net/url"
	"strconv"
	"time"
)

const lastCheckTimeFormat = "2006-01-02 15:04:05"

type nhl struct {
	client *gonhl.Client
}

func InitNHL() *nhl {
	return &nhl{
		client: gonhl.NewClient(),
	}
}

func (n *nhl) Name() string {
	return "nhl"
}

func (n *nhl) Schedule(params url.Values) map[string]interface{} {
	schedule, _ := n.client.GetSchedule(buildScheduleParamsFromParams(params))
	return buildResultFromNHLSchedule(&schedule)
}

func (n *nhl) PlayByPlay(params url.Values) map[string]interface{} {
	id, _ := strconv.Atoi(params.Get("gameId"))
	liveData, _ := n.client.GetGameLiveData(id)
	lastCheckString := params.Get("date")
	result := buildResultFromLiveData(&liveData, lastCheckString)
	return result
}

func (n *nhl) ParseScheduleState(statusCode int) ScheduleState {
	status := Preview
	if statusCode >= 5 { // 5 6 7 states
		status = Complete
	} else if statusCode > 2 { // 3 4 states
		status = Live
	}
	return status
}

func (n *nhl) DefaultTimeString() string {
	return time.Now().In(time.UTC).Format(lastCheckTimeFormat)
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

func buildResultFromLiveData(liveData *gonhl.LiveData, lastCheckString string) map[string]interface{} {
	lastCheck, _ := time.Parse(lastCheckTimeFormat, lastCheckString)
	result := make(map[string]interface{})
	result["game"] = buildGameFromLiveData(liveData)
	result["plays"] = buildPlaysFromPlays(&liveData.Plays, &lastCheck)
	result["players"] = buildPlayersFromBoxScore(&liveData.Boxscore)
	metadata := make(map[string]interface{})
	metadata["state"] = buildStateFromLiveData(liveData)
	metadata["lastCheck"] = liveData.Plays.CurrentPlay.About.DateTime.Format(lastCheckTimeFormat)
	result["metadata"] = metadata
	return result
}

func buildGameFromLiveData(liveData *gonhl.LiveData) map[string]interface{} {
	game := make(map[string]interface{})
	game["home"] = buildTeamFromLinescore(&liveData.Linescore.Teams.Home)
	game["away"] = buildTeamFromLinescore(&liveData.Linescore.Teams.Away)
	game["status"] = buildStatusFromLinescore(&liveData.Linescore)
	return game
}

func buildStateFromLiveData(liveData *gonhl.LiveData) ScheduleState {
	resultEvent := liveData.Plays.CurrentPlay.Result.EventTypeID
	if resultEvent == "GAME_OFFICIAL" || resultEvent == "GAME_END" {
		return Complete
	} else if resultEvent == "PERIOD_OFFICIAL" || resultEvent == "PERIOD_END" {
		return Intermission
	}
	return Live
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

func buildPlaysFromPlays(playsData *gonhl.Plays, lastCheck *time.Time) []map[string]interface{} {
	plays := make([]map[string]interface{}, 0, len(playsData.AllPlays))
	for _, playData := range playsData.AllPlays {
		if lastCheck == nil || playData.About.DateTime.Sub(*lastCheck) > 0 {
			play := make(map[string]interface{})
			play["description"] = playData.Result.Description
			play["typeId"] = playData.Result.EventTypeID
			play["periodTime"] = playData.About.PeriodTime
			play["coordinates"] = map[string]float64{"x": playData.Coordinates.X, "y": playData.Coordinates.Y}
			play["dateTime"] = playData.About.DateTime.Format(lastCheckTimeFormat)
			plays = append(plays, play)
		}
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
	result := make([]map[string]interface{}, len(team.OnIcePlus))
	for i, onIceSkater := range team.OnIcePlus {
		skater := make(map[string]interface{})
		stringSkaterId := fmt.Sprintf("ID%d", onIceSkater.PlayerID)
		skaterData := team.Players[stringSkaterId]
		skater["id"] = skaterData.Person.ID
		skater["onIceDuration"] = onIceSkater.ShiftDuration
		skater["name"] = skaterData.Person.FullName
		skater["number"] = skaterData.JerseyNumber
		skater["position"] = skaterData.Position.Abbreviation

		result[i] = skater
	}
	return result
}

func buildResultFromNHLSchedule(schedule *gonhl.Schedule) map[string]interface{} {
	result := make(map[string]interface{})
	result["content"] = []string{}
	if len(schedule.Dates) == 0 {
		return result
	}
	resultGames := make([]map[string]interface{}, schedule.Dates[0].TotalGames)
	games := schedule.Dates[0].Games
	for i, game := range games {
		resultGame := make(map[string]interface{})
		resultGame["id"] = game.GamePk
		resultGame["date"] = CreateDetailedStringFromDate(game.GameDate)
		resultGame["status"] = game.Status.AbstractGameState
		resultGame["statusCode"] = game.Status.CodedGameState
		resultGame["period"] = game.Linescore.CurrentPeriod
		resultGame["time"] = game.Linescore.CurrentPeriodTimeRemaining
		resultGame["home"] = parseNHLTeam(game.Teams.Home)
		resultGame["away"] = parseNHLTeam(game.Teams.Away)
		resultGame["venue"] = game.Venue.Name
		resultGames[i] = resultGame
	}
	result["content"] = resultGames
	return result
}

func parseNHLTeam(team gonhl.GameTeam) map[string]interface{} {
	resultTeam := make(map[string]interface{})
	resultTeam["teamId"] = team.Team.ID
	resultTeam["name"] = team.Team.Name
	resultTeam["abbr"] = team.Team.ShortName
	resultTeam["record"] = fmt.Sprintf("%d-%d-%d", team.LeagueRecord.Wins, team.LeagueRecord.Losses, team.LeagueRecord.Ot)
	resultTeam["score"] = team.Score
	return resultTeam
}

