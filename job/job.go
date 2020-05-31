package job

import (
	"errors"
	"fmt"
	"github.com/go-pg/pg"
	"log"
	"sync"
	"time"
)

type Job struct {
	sync.Mutex
	ID              *int64    `json:"id" sql:"id"`
	Frequency       string    `json:"frequency,omitempty" sql:"frequency"`
	StartDateString string    `json:"start_date,omitempty" sql:"-"`
	EndDateString   string    `json:"end_date,omitempty" sql:"-"`
	StartDate       time.Time `json:"start_date_time" sql:"start_date"`
	EndDate         time.Time `json:"end_date_time" sql:"end_date"`
	CronEntryID     int     `json:"-" sql:"cron_entry_id"`
}

func (j Job) Run() {
	if time.Now().UnixNano() > j.StartDate.UnixNano() {
		// run script
		log.Println("Hi from job: ",j.ID)
	} else if time.Now().UnixNano() > j.EndDate.UnixNano() {
		fmt.Println("stoping")
	}
}

func (j *Job) FormatJobData() (err error) {
	j.StartDate, err = time.Parse("02-01-2006", j.StartDateString)
	if err != nil {
		log.Println(err)
		err = errors.New("invalid start date, date format: dd-mm-yyyy")
		return
	}
	j.EndDate, err = time.Parse("02-01-2006", j.EndDateString)
	if err != nil {
		log.Println(err)
		err = errors.New("invalid end date, date format: dd-mm-yyyy")
		return
	}
	if j.EndDate.Before(time.Now()) {
		log.Println("End date: ", j.EndDate)
		err = errors.New("invalid end_date")
		return
	}
	if j.Frequency==""{
		log.Println("invalid frequency: ", j.Frequency)
		err = errors.New("invalid frequency")
		return
	}
	return
}

func (j *Job) List(db *pg.DB) ([]Job, error) {
	var list []Job
	err := db.Model(&list).Select()
	return list, err
}
func (j *Job) FetchJob(db *pg.DB) error {
	err := db.Model(j).WherePK().Select()
	return err
}
func (j *Job) Store(db *pg.DB) error {
	j.ID = nil
	err := db.Insert(j)
	return err
}

func (j *Job) Delete(db *pg.DB) error {
	_, err := db.Model(j).WherePK().Delete()
	return  err
}

func (j *Job) Update(db *pg.DB) error {
	_, err := db.Model(j).Set("start_date =?start_date").Set("end_date =?end_date").Set("frequency =?frequency").
		WherePK().Update()
	return  err
}
