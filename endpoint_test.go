package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/taalhach/restcron/job"
	"github.com/taalhach/restcron/server"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

const (
	serverUrl="http://localhost:8080"
		)
func TestJob(t *testing.T)  {

	//to use mock server uncomment following lines but its is recommended to run server separately for clean testing log

	//mockServer()
	//log.Print("\n\n\n**************************** Starting test *******************\n\n\n")
	var id int64
	t.Run(fmt.Sprintf("Get_all_jobs"), func(t *testing.T) {
		req,err:=http.NewRequest(http.MethodGet,serverUrl+"/",nil)
		if err!=nil{
			t.Error(err)
			return
		}
		client:=http.Client{}
		resp,err:= client.Do(req)
		if err!=nil{
			t.Error(err)
			return
		}
		if resp.StatusCode!=http.StatusOK{
			t.Error("status code: ",resp.StatusCode)
		}
	})
	t.Run(fmt.Sprintf("Create_Job"), func(t *testing.T) {
		body,err:=json.Marshal(&job.Job{EndDate:time.Now().Add(time.Minute),StartDate:time.Now(),Frequency:"@every 3s"})
		if err!=nil{
			t.Error(err)
			return
		}
		req,err:=http.NewRequest(http.MethodPost,serverUrl+"/job",bytes.NewBuffer(body))
		if err!=nil{
			t.Error(err)
			return
		}
		client:=http.Client{}
		resp,err:= client.Do(req)
		if err!=nil{
			t.Error(err)
			return
		}
		if resp.StatusCode!=http.StatusOK{
			t.Error("status code: ",resp.StatusCode)
			return
		}
		body,err=ioutil.ReadAll(resp.Body)
		if err!=nil{
			t.Error(err)
			return
		}
		var j job.Job
		if err:=json.Unmarshal(body,&j); err!=nil{
			t.Error(err)
			return
		}
		if j.ID==nil  {
			t.Error("Job not created",j.ID)
			return
		}
		id=*j.ID
	})
	t.Run(fmt.Sprintf("Update_Job/%v",id), func(t *testing.T) {
		body,err:=json.Marshal(&job.Job{ID:&id,EndDate:time.Now().Add(time.Minute),StartDate:time.Now(),Frequency:"@every 3s"})
		if err!=nil{
			t.Error(err)
			return
		}
		req,err:=http.NewRequest(http.MethodPatch,serverUrl+"/job",bytes.NewBuffer(body))
		if err!=nil{
			t.Error(err)
			return
		}
		client:=http.Client{}
		resp,err:= client.Do(req)
		if err!=nil{
			t.Error(err)
			return
		}
		if resp.StatusCode!=http.StatusOK{
			t.Error("status code: ",resp.StatusCode)
			return
		}
		body,err=ioutil.ReadAll(resp.Body)
		if err!=nil{
			t.Error(err)
			return
		}
		var j job.Job
		if err:=json.Unmarshal(body,&j); err!=nil{
			t.Error(err)
			return
		}
		if j.ID==nil  {
			t.Error("Job not updated",j.ID)
			return
		}
	})
	t.Run(fmt.Sprintf("Get_Job/%v",id), func(t *testing.T) {
		req,err:=http.NewRequest(http.MethodGet,serverUrl+fmt.Sprintf("/job/%v",id),nil)
		if err!=nil{
			t.Error(err)
			return
		}
		client:=http.Client{}
		resp,err:= client.Do(req)
		if err!=nil{
			t.Error(err)
			return
		}
		if resp.StatusCode!=http.StatusOK{
			t.Error("status code: ",resp.StatusCode)
		}
	})
	t.Run(fmt.Sprintf("Delete Job/%v",id), func(t *testing.T) {
		req,err:=http.NewRequest(http.MethodDelete,serverUrl+fmt.Sprintf("/job/%v",id),nil)
		if err!=nil{
			t.Error(err)
			return
		}
		client:=http.Client{}
		resp,err:= client.Do(req)
		if err!=nil{
			t.Error(err)
			return
		}
		if resp.StatusCode!=http.StatusOK{
			t.Error("status code: ",resp.StatusCode)
			return
		}
		body,err:=ioutil.ReadAll(resp.Body)
		if err!=nil{
			t.Error(err)
			return
		}
		var j job.Job
		if err:=json.Unmarshal(body,&j); err!=nil{
			t.Error(err)
			return
		}
		if j.ID==nil  {
			t.Error(fmt.Sprintf("Job %v was supposed to delete but deleted: %v",id,j.ID))
			return
		}else if *j.ID!=id || j.IsActive{
			t.Error(fmt.Sprintf("Job %v was supposed to delete but status is: %v",id,*j.ID))
			return
		}
	})
}
func mockServer()  {
	//mock server is started in a go routine
	go func() {
		cfg,err:=readConfig()
		if err!=nil{
			log.Fatal(err)
		}
		s, err := server.NewServer(cfg.Database.Url, cfg.Database.User_name, cfg.Database.Password)
		if err != nil {
			log.Fatal(err)
		}
		s.RunServer(cfg.Server.Port)
	}()
	time.Sleep(3*time.Second)
}
