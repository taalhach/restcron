package main

import (
	"github.com/taalhach/restcron/server"
	"log"
)

//func newJob(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	cron.New()
//	reqBody, _ := ioutil.ReadAll(r.Body)
//	var j job.Job
//	err:=json.Unmarshal(reqBody, &j)
//	if err!=nil{
//		log.Println(err)
//		return
//	}
//	err=j.FormatDate()
//	if err!=nil{
//		w.Write([]byte(err.Error()))
//		return
//	}
//	fmt.Println(j)
//	db:=pg.Connect(&pg.Options{User:"postgres",Password:"postres"})
//	_,err=db.Exec("SELECT 1")
//	fmt.Println(err)
//	fmt.Println(j.List(db))
//	//j.Store(db)
//	//j.List(db)
//	//j.Delete(db)
//	// add to database
//	// add to map for after checking
//
//}

func main() {
	s, err := server.NewServer("postgres", "postgres")
	if err != nil {
		log.Fatal(err)
	}
	s.RunServer()
	//router := httprouter.New()
	//router.POST("/",newJob)
	//http.ListenAndServe(":8080",router)
}
