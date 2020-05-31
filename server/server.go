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

//gives instance of server
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

// starts jobs remover,loads router and starts server
func (s *HttpServer) RunServer() {
	table:="create table if not exists jobs (id serial primary key, start_date timestamptz, end_date timestamptz,frequency varchar(255),cron_entry_id integer);"
	_, err := s.db.Exec(table, nil)
	if err != nil {
		panic(err)
	}
	s.recover()
	go s.jobRemover()
	s.loadRoutes()
	s.cron.Start()
	http.ListenAndServe(":8080", s.router)
}

// this function removes the job that is expired
func (s *HttpServer) jobRemover() {
	for id:=range job.RemoveJob{
		s.cron.Remove(cron.EntryID(id))
		log.Printf("entry %v is expired and removed \n",id)
	}
}
// recovers jobs when server starts by getting jobs from database and removes if expired
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

// loads router
func (s *HttpServer) loadRoutes() {
	s.router.GET("/", s.getAllJobs())
	s.router.GET("/job/:id", s.getJob())
	s.router.POST("/job", s.crateJob())
	s.router.PATCH("/job", s.updateJob())
	s.router.DELETE("/job/:id", s.removeJob())
}
