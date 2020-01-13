# GoSports API Server

## Endpoints

### Schedule 

url: `/schedule/{sport}`

- sport: [mlb, nba, nfl, nhl]

parameters:

- date (_Optional_): [yyyy-mm-dd]

returns:

- array of games:

example:

````
{
    content: [
        {
            id: <string>,
            date: <2006-01-02T15:04:05Z07:00>,
            status: [Scheduled, Active, Complete],
            statusCode: <int>,
            period: <int>,
            time: <string>, //Time remaining in period/quarter, or number of outs in the inning
            home: {
                teamId: <string>,
                name: <string>,
                abbr: <string>,
                record: <string>,
                score: <int>
            },
            away: {
                teamId: <string>,
                name: <string>,
                abbr: <string>,
                record: <string>,
                score: <int>
            },
            venue: <string>
        }
    ]
}
````

### Play By Play

url: `/playbyplay/{sport}`

- sport: [mlb, nba, nfl, nhl]

parameters:

- gameId (_Required_):

returns (internal structure changes based on sport): 

- base information (diff)

- array of plays (diff)

example:

````
{
    game: {
        status: {},
        home: {},
        away: {}
    }, 
    players: {
        home: [],
        away: []
    },
    plays: [],
    metadata: {
        state: <int>,
        lastCheck: <string>,
    }
}
````

#### NHL 

````
{
    game: {
        status: {
            period: <int>,
            periodTimeRemaining: <int>
        },
        home: {
            name: <string>,
            score: <string>,
            shots: <int>
        },
        away: {
            name: <string>,
            score: <string>,
            shots: <int>
        }
    }, 
    players: {
        home: [
            {
                id: <int>,
                onIceDuration: <int>,
                name: <string>,
                number: <string>,
                position: <string>
            }, ...
        ],
        away: [
            {
                id: <int>,
                onIceDuration: <int>,
                name: <string>,
                number: <string>,
                position: <string>
            }, ...
        ]
    }
    plays: [
        {
            description: <string>,
            typeId: <string>
            periodTime: <string>
            dateTime: <string>
            coordinates: {
                x: <int>,
                y: <int>
            }
        }, ...
    ],
    metadata: {
        state: <int>,
        lastCheck: <string>,
    }
}
````

## Official API Documentation

### NHL

Base: https://statsapi.web.nhl.com/api/v1/

- Schedule:

    - Endpoint: `schedule`
    
    - Parameters: [expand (schedule.broadcasts, schedule.linescore, schedule.ticket), teamId, date (yyyy-mm-dd), startDate, endDate]
    
    - Example: https://statsapi.web.nhl.com/api/v1/schedule?expand=schedule.broadcasts,schedule.linescore&teamId=30

- Play by Play: 

    - Endpoint: `game/%d/feed/live`
    
    - Parameters: [gamePk]
    
    - Example: https://statsapi.web.nhl.com/api/v1/game/2019020195/feed/live
    
- Play by Play Diff: 

    - Endpoint: `game/%d/feed/live/diffPatch`
    
    - Parameters: [gamePk, startTimecode (yyyymmdd_hhmmss)]
    
    - Example: https://statsapi.web.nhl.com/api/v1/game/2018020150/feed/live/diffPatch?startTimecode=20181027_1600

### NBA

Base: https://stats.nba.com/stats/

BaseV2: http://data.nba.com/data/5s/json/cms/noseason/

- Schedule:

    - Endpoint: `scoreboardv2`
    
    - Parameters: [GameDate (yyyy-mm-dd), LeagueID, DayOffset]
    
    - Example: https://stats.nba.com/stats/scoreboardv2/?GameDate=2019-10-31&LeagueID=00&DayOffset=0

- Play by Play: 

    - Endpoint: `playbyplayv2`
    
    - Parameters: [GameID, StartPeriod (1 - 4), EndPeriod]
    
    - Example: https://stats.nba.com/stats/playbyplayv2/?GameID=0021900054&StartPeriod=1&EndPeriod=4
    
- ScheduleV2: 

    - Endpoint: `scoreboard/%s/games.json`
    
    - Parameters: [Date (yyyymmdd)]
    
    - Example: http://data.nba.com/data/5s/json/cms/noseason/scoreboard/20191120/games.json

- Play by Play V2:

    - Endpoint: `game/%s/%s/pbp_all.json`
    
    - Parameters: [Date (yyyymmmdd), GameID]
    
    - Example: http://data.nba.com/data/5s/json/cms/noseason/game/20191118/0021900189/pbp_all.json
    
### NFL

Base: https://www.nfl.com/

- Schedule:

    - Endpoint: `ajax/scorestrip` && `feeds-rs/currentWeek.json`
    
    - Parameters: [season (yyyy), seasonType (PRE, REG, POST), week]
    
    - Example: https://www.nfl.com/ajax/scorestrip/?season=2019&seasonType=REG&week=9 && https://www.nfl.com/feeds-rs/currentWeek.json

- Play by Play: 

    - Endpoint: `liveupdate/game-center/%s/%s_gtd.json`
    
    - Parameters: [eid, eid]
    
    - Example: https://www.nfl.com/liveupdate/game-center/2019103100/2019103100_gtd.json
    
### MLB

Base: https://statsapi.mlb.com/api/v1.1/

- Schedule (API v1):

    - Endpoint: `schedule`
    
    - Parameters: [scheduleType, eventTypes, hydrate (decisions, probablePitcher(note), linescore), teamId, leagueId, sportId, gamePk, gamePks, venueIds, gameTypes, date (yyyy-mm-dd), startDate, endDate, opponentId, fields]
    
    - Example: http://statsapi.mlb.com/api/v1/schedule/?sportId=1&date=2019-10-30&hydrate=team

- Play by Play (API v1.1): 

    - Endpoint: `game/%s/feed/live`
    
    - Parameters: [gamePk]
    
    - Example: https://statsapi.mlb.com/api/v1.1/game/599377/feed/live