package main

func (s *server) routes() {
	s.router.HandleFunc("/about", s.handleAbout())
	s.router.HandleFunc("/schedule/{sport}", s.handleSchedule())

	s.router.HandleFunc("/client/{sport}", s.checkValidQueries(s.websocketUpgrade(s.handlePlayByPlay()),
		[]ValidateParameter{parseSport},
		[]ValidateQuery{parseGameId}))
}
