package server

import (
	"github.com/go-pg/pg"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron/v3"
	"net/http"
)


type HttpServer struct {
	db *pg.DB
	router *httprouter.Router
	cron *cron.Cron
}

func NewServer(User,Password string) (*HttpServer,error) {
	db:=pg.Connect(&pg.Options{
		Password:Password,
		User:User,
	})
	_,err:=db.Exec("SELECT 1")
	if err!=nil{
		return nil,err
	}
	cron:=cron.New()
	return &HttpServer{
		db:   db,
		cron: cron,
		router:httprouter.New(),
	},nil
}

func (s *HttpServer)RunServer()  {
	s.loadRoutes()
	s.cron.Start()
	http.ListenAndServe(":8080",s.router)

}

func (s *HttpServer)loadRoutes()  {
	s.router.GET("/",s.getAllJobs())
	s.router.GET("/job/:id",s.getJob())
	s.router.POST("/job",s.crateJob())
	s.router.PATCH("/job",s.updateJob())
	s.router.DELETE("/job/:id",s.removeJob())
}