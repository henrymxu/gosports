package main

func (s *server) routes() {
	s.router.HandleFunc("/about", s.handleAbout())
	s.router.HandleFunc("/schedule/{sport}", s.handleSchedule())
	//s.router.HandleFunc("/playbyplay/{sport}", s.handlePlayByPlay())

	pbpParamsCheck := []ValidateParameter{parseSport}
	pbpQueriesCheck := []ValidateQuery{parseGameId}
	s.router.HandleFunc("/client/{sport}", s.checkValidQueries(s.websocketUpgrade(s.handlePlayByPlay()),
			pbpParamsCheck,
			pbpQueriesCheck))
}
