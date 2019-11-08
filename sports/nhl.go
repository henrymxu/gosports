package sports

import (
	"fmt"
	"github.com/henrymxu/gonhl"
	"net/url"
)

type nhl struct {
	client *gonhl.Client
}

func (n nhl) Schedule(params url.Values) map[string]interface{} {
	if n.client == nil {
		n.client = gonhl.NewClient()
	}
	schedule, _ := n.client.GetSchedule(buildScheduleParamsFromParams(params))
	return buildResultFromSchedule(schedule)
}

func (n nhl) PlayByPlay(params url.Values) map[string]interface{} {
	if n.client == nil {
		n.client = gonhl.NewClient()
	}

	return nil
}

func buildScheduleParamsFromParams(params url.Values) *gonhl.ScheduleParams {
	scheduleParams := gonhl.NewScheduleParams()
	scheduleParams.ShowLinescore()
	val := string(params.Get("date"))
	if val != "" {
		date, err := CreateDateFromString(val)
		if err == nil {
			scheduleParams.SetDate(date)
		}
	}
	return scheduleParams
}

func buildResultFromPlays(plays gonhl.Plays) map[string] interface{} {
	for _, play := range plays.AllPlays {

	}
}

func buildResultFromSchedule(schedule gonhl.Schedule) map[string]interface{} {
	result := make(map[string]interface{})
	result["content"] = []string{}
	if len(schedule.Dates) == 0 {
		return result
	}
	resultGames := make([]map[string]interface{}, 0, schedule.Dates[0].TotalGames)
	games := schedule.Dates[0].Games
	for _, game := range games {
		resultGame := make(map[string]interface{})
		resultGame["gameId"] = game.GamePk
		resultGame["date"] = CreateStringFromDate(game.GameDate)
		resultGame["status"] = game.Status.AbstractGameState
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

