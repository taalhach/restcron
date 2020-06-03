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

const (
	jobTableSchema=`create table  if not exists jobs (id serial primary key,
					start_date timestamptz, end_date timestamptz,
					frequency varchar(255),cron_entry_id integer,
					is_active bool default false);`
)

type HttpServer struct {
	db     *pg.DB
	router *httprouter.Router
	cron   *cron.Cron
}

//gives instance of server
func NewServer(addr,user, password string) (*HttpServer, error) {
	db := pg.Connect(&pg.Options{
		Addr: addr,
		Password: password,
		User:     user,
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
func (s *HttpServer) RunServer(serverPort string) {
	_, err := s.db.Exec(jobTableSchema, nil)
	if err != nil {
		panic(err)
	}
	s.recover()
	go s.jobRemover()
	s.loadRoutes()
	s.cron.Start()
	log.Println("Server started")
	http.ListenAndServe(":"+serverPort, s.router)
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
	log.Println("recovering from database if any job exists and not expired")
	_job := job.Job{}
	_jobs, err := _job.List(s.db)
	if err != nil {
		log.Fatal(err)
	}
	for _,_job = range _jobs {
		 if (_job.EndDate.Unix()<0 || _job.EndDate.UnixNano() > time.Now().UnixNano()) && _job.IsActive {
			entryID, err := s.cron.AddJob(_job.Frequency, _job)
			if err != nil {
				log.Println(err)
			}
			_job.CronEntryID = int(entryID)
			err = _job.Update(s.db)
			if err != nil {
				log.Println(err)
			}
		}else if _job.CronEntryID>0 {
			_job.CronEntryID=0
			err = _job.Update(s.db)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// loads router
func (s *HttpServer) loadRoutes() {
	s.router.GET("/", middleware(s.getAllJobs()))
	s.router.GET("/job/:id", middleware(s.getJob()))
	s.router.POST("/job", middleware(s.crateJob()))
	s.router.PATCH("/job", middleware(s.updateJob()))
	s.router.DELETE("/job/:id", middleware(s.removeJob()))
}

func middleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println(r.URL.String(), "	", r.Method, "		")
		next(w, r, p)
	}
}


