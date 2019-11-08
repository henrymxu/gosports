package main

func (s *server) routes() {
	s.router.HandleFunc("/about", s.handleAbout())
	s.router.HandleFunc("/schedule/{sport}", s.handleSchedule())
	//s.router.HandleFunc("/playbyplay/{sport}", s.handlePlayByPlay())
	s.router.HandleFunc("/client/{sport}", s.websocketUpgrade(s.handlePlayByPlay()))
}
