package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron/v3"
	"github.com/taalhach/restcron/job"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)
type _error struct {
	Errors interface{} `json:"errors"`
}

func (s *HttpServer) getAllJobs() func(w http.ResponseWriter,r *http.Request,_ httprouter.Params){
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_job:=&job.Job{}
		jobs,err:=_job.List(s.db)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"internal server error"},
			},http.StatusInternalServerError,w)
			return
		}
		writeResponse(struct {
			Jobs []job.Job `json:"jobs"`
		}{
			Jobs:jobs,
		},http.StatusOK,w)
	}
}
func (s *HttpServer) getJob() func(w http.ResponseWriter,r *http.Request,_ httprouter.Params){
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		n:=ps.ByName("id")
		id,err:=strconv.ParseInt(n,10,64)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"invalid job id"},
			},http.StatusBadRequest,w)
			return
		}
		_job:=&job.Job{ID:&id}
		err=_job.FetchJob(s.db)
		if err!=nil{
			log.Println(err)
			_code:=http.StatusInternalServerError
			_resp:=_error{
				Errors:[]string{"internal server error"},
			}
			if strings.Contains(err.Error(),"no rows"){
				log.Printf("Job: %v is not found",*_job.ID)
				_resp.Errors=[]string{"job not found\n"}
				_code=http.StatusNotFound
			}
			writeResponse(_resp,_code,w)
			return
		}
		writeResponse(_job,http.StatusOK,w)
	}
}
func (s *HttpServer) crateJob() func(w http.ResponseWriter,r *http.Request,_ httprouter.Params){
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		_job :=&job.Job{}
		err:=json.Unmarshal(reqBody, _job)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"bad request"},
			},http.StatusBadRequest,w)
			return
		}
		err=_job.FormatJobData()
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{err.Error()},
			},http.StatusBadRequest,w)
			return
		}
		if _job.IsActive {
			//create cron job
			entryID,err:=s.cron.AddJob(_job.Frequency,_job)
			if err!=nil{
				log.Println(err)
				writeResponse(_error{
					Errors:[]string{err.Error()},
				},http.StatusInternalServerError,w)
				return
			}
			// add this entry id to job and store in database
			_job.CronEntryID=int(entryID)
		}
		err=_job.Store(s.db)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"internal server error"},
			},http.StatusInternalServerError,w)
			return
		}
		writeResponse(_job,http.StatusOK,w)
		log.Printf("Job is created successfully, JobID: %v, Cron_entry_id: %v, Frequency: %v, Start_date: %v, End_date: %v\n",
			*_job.ID,_job.CronEntryID,_job.Frequency,_job.StartDate,_job.EndDate.Unix())
	}
}
func (s *HttpServer) removeJob() func(w http.ResponseWriter,r *http.Request,_ httprouter.Params){
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		n:=ps.ByName("id")
		id,err:=strconv.ParseInt(n,10,64)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"invalid job id"},
			},http.StatusBadRequest,w)
			return
		}
		_job:=&job.Job{ID:&id}
		err=_job.FetchJob(s.db)
		if err!=nil{
			log.Println(err)
			_code:=http.StatusInternalServerError
			_resp:=_error{
				Errors:[]string{"internal server error"},
			}
			if strings.Contains(err.Error(),"no rows"){
				log.Printf("Job: %v is not found\n",*_job.ID)
				_resp.Errors=[]string{"job not found"}
				_code=http.StatusNotFound
			}
			writeResponse(_resp,_code,w)
			return
		}
		_entry:=cron.EntryID(_job.CronEntryID)
		s.cron.Remove(_entry)
		_job.IsActive=false
		_job.CronEntryID=0
		err=_job.Update(s.db)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"internal server error"},
			},http.StatusInternalServerError,w)
			return
		}
		writeResponse(_job,http.StatusOK,w)
		log.Printf("Job is paused successfully, JobID: %v, Cron_entry_id: %v, Frequency: %v, Start_date: %v, End_date: %v\n",
			*_job.ID,_job.CronEntryID,_job.Frequency,_job.StartDate,_job.EndDate)

	}
}
func (s *HttpServer) updateJob() func(w http.ResponseWriter,r *http.Request,_ httprouter.Params){
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		_job :=&job.Job{}
		err:=json.Unmarshal(reqBody, _job)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"bad request"},
			},http.StatusBadRequest,w)
			return
		}
		err=_job.FormatJobData()
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{err.Error()},
			},http.StatusBadRequest,w)
			return
		}
		// create an instance to get entry id from database
		_dbJob:=&job.Job{ID:_job.ID}
		err=_dbJob.FetchJob(s.db)
		if err!=nil{
			log.Println(err)
			_code:=http.StatusInternalServerError
			_resp:=_error{
				Errors:[]string{"internal server error"},
			}
			if strings.Contains(err.Error(),"no rows"){
				log.Printf("Job: %v is not found\n",*_job.ID)
				_resp.Errors=[]string{"job not found"}
				_code=http.StatusNotFound
			}
			writeResponse(_resp,_code,w)
			return
		}
		_entry:=cron.EntryID(_dbJob.CronEntryID)
		// both status are set then change to new properties
		if _job.IsActive && _dbJob.IsActive{
			// remove that cron job
			s.cron.Remove(_entry)
			// add new cron job
			_entry,err=s.cron.AddJob(_job.Frequency,_job)
			if err!=nil{
				log.Println(err)
				writeResponse(_error{
					Errors:[]string{err.Error()},
				},http.StatusInternalServerError,w)
				return
			}
			// add this cron new entry id
			_job.CronEntryID=int(_entry)

		} else if _dbJob.IsActive && !_job.IsActive{
			// new status is off then remove the job from cron
				s.cron.Remove(_entry)
				_job.CronEntryID=0

		}else if !_dbJob.IsActive && _job.IsActive   {
			// add new cron job when new status is on
			_entry,err=s.cron.AddJob(_job.Frequency,_job)
			if err!=nil{
				log.Println(err)
				writeResponse(_error{
					Errors:[]string{err.Error()},
				},http.StatusInternalServerError,w)
				return
			}
			// add this cron new entry id
			_job.CronEntryID=int(_entry)
		}
		err=_job.Update(s.db)
		if err!=nil{
			log.Println(err)
			writeResponse(_error{
				Errors:[]string{"internal server error"},
			},http.StatusInternalServerError,w)
			return
		}
		writeResponse(_job,http.StatusOK,w)
		log.Printf("Job is updated successfully, JobID: %v, Cron_entry_id: %v, Frequency: %v, Start_date: %v, End_date: %v\n",
			*_job.ID,_job.CronEntryID,_job.Frequency,_job.StartDate,_job.EndDate)

	}
}

func writeResponse(resp interface{},code int, w http.ResponseWriter) {
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
