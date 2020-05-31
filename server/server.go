package server

import (
	"github.com/go-pg/pg"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron/v3"
	"github.com/taalhach/restcron/job"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	db     *pg.DB
	router *httprouter.Router
	cron   *cron.Cron
}

func NewServer(User, Password string) (*HttpServer, error) {
	db := pg.Connect(&pg.Options{
		Password: Password,
		User:     User,
	})
	_, err := db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}
	cron := cron.New()
	return &HttpServer{
		db:     db,
		cron:   cron,
		router: httprouter.New(),
	}, nil
}

func (s *HttpServer) RunServer() {
	s.recover()
	s.loadRoutes()
	s.cron.Start()
	http.ListenAndServe(":8080", s.router)

}

func (s *HttpServer) recover() {
	_job := job.Job{}
	_jobs, err := _job.List(s.db)
	if err != nil {
		log.Fatal(err)
	}
	for _,_job = range _jobs {
		if _job.EndDate.UnixNano() < time.Now().UnixNano() {
			err = _job.Delete(s.db)
			if err != nil {
				log.Println(err)
			}
		} else {
			entryID, err := s.cron.AddJob(_job.Frequency, _job)
			if err != nil {
				log.Println(err)
			}
			_job.CronEntryID = int(entryID)
			err = _job.Update(s.db)
			if err != nil {
				log.Println(err)
			}
		}

	}
}

func (s *HttpServer) loadRoutes() {
	s.router.GET("/", s.getAllJobs())
	s.router.GET("/job/:id", s.getJob())
	s.router.POST("/job", s.crateJob())
	s.router.PATCH("/job", s.updateJob())
	s.router.DELETE("/job/:id", s.removeJob())
}
