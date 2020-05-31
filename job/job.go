package job

import (
	"errors"
	"github.com/go-pg/pg"
	"log"
	"sync"
	"time"
)

type Job struct {
	sync.Mutex
	ID              *int64    `json:"id" sql:"id"`
	Frequency       string    `json:"frequency,omitempty" sql:"frequency"`
	StartDate       time.Time `json:"start_date" sql:"start_date"`
	EndDate         time.Time `json:"end_date" sql:"end_date"`
	CronEntryID     int     `json:"-" sql:"cron_entry_id"`
}

func (j Job) Run() {
	if time.Now().UnixNano() > j.StartDate.UnixNano() && time.Now().UnixNano() < j.EndDate.UnixNano() {
		// run script
		log.Println("Hi from job: ",*j.ID)
	} else if time.Now().UnixNano() > j.EndDate.UnixNano() {
		//remove job from cron automatically
	}
}

func (j *Job) FormatJobData() (err error) {
	if j.EndDate.Before(time.Now()) {
		log.Println("End date: ", j.EndDate)
		err = errors.New("invalid end_date")
		return
	}
	j.EndDate.In(time.Now().Location())
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
