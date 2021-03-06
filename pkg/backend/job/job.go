package job

import (
	"github.com/RedHatGov/ocdb/pkg/utils"
	"github.com/gobuffalo/buffalo/worker"
	"time"
)

type JobFn func() error

type Job struct {
	Name         string
	Fn           JobFn `json:"-"`
	Period       time.Duration
	DelayedStart time.Duration
	LastStart    time.Time
	LastSuccess  time.Time
	LastError    string
	ErrorCount   uint
}

func (job *Job) SetUpIn(w worker.Worker) {
	err := w.Register(job.Name, func(args worker.Args) error {
		job.reschedule(w, job.Period)
		return job.run()
	})
	if err != nil {
		utils.Log.Fatalf("Could not register job: %v", err)
	}
	job.reschedule(w, job.DelayedStart+time.Second)
}

func (job *Job) reschedule(w worker.Worker, period time.Duration) {
	err := w.PerformIn(worker.Job{
		Queue:   "masonry",
		Handler: job.Name,
	}, period)
	if err != nil {
		utils.Log.Errorf("Could not reschedule job %s: %v", job.Name, err)
	}
}

func (job *Job) run() error {
	job.LastStart = time.Now()

	err := job.Fn()
	if err != nil {
		job.LastError = err.Error()
		job.ErrorCount += 1
		if job.ErrorCount < 16 {
			utils.Log.Errorf("Job '%s' Failed: %s. Re-trying", job.Name, job.LastError)
			return job.run()
		}
	} else {
		job.LastError = ""
		job.LastSuccess = time.Now()
		job.ErrorCount = 0
	}
	return err
}
