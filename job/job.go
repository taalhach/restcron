package job

import (
	"errors"
	"github.com/go-pg/pg"
	"log"
	"sync"
	"time"
)

var RemoveJob = make(chan int)

type Job struct {
	sync.Mutex
	ID          *int64    `json:"id" sql:"id"`
	Frequency   string    `json:"frequency,omitempty" sql:"frequency"`
	StartDate   time.Time `json:"start_date,omitempty" sql:"start_date"`
	EndDate     time.Time `json:"end_date,omitempty" sql:"end_date"`
	CronEntryID int       `json:"-" sql:"cron_entry_id"`
	IsActive    bool      `json:"is_active" sql:"is_active"`
}

func (j Job) Run() {
	if time.Now().Unix() > j.StartDate.Unix()&& ((j.EndDate.Unix()>0 && j.EndDate.Unix()>time.Now().Unix())||j.EndDate.Unix()<0){
		log.Println("Hi from job: ", *j.ID)

	} else if j.EndDate.UnixNano()>0 && time.Now().UnixNano() > j.EndDate.UnixNano() {
		//inform entry id of job to remove job from cron when it is expired
		RemoveJob <- j.CronEntryID
		log.Printf("job %v is expired ", *j.ID)
	}
}

func (j *Job) FormatJobData() (err error) {
	if j.Frequency == "" {
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
	_, err := db.Model(j).
		Set("is_active =false").
		WherePK().
		Update()
	return err
}

func (j *Job) Update(db *pg.DB) error {
	_, err := db.Model(j).
		Set("start_date =?start_date").
		Set("end_date =?end_date").
		Set("frequency =?frequency").
		Set("cron_entry_id =?cron_entry_id").
		Set("is_active =?is_active").
		WherePK().
		Update()
	return err
}
