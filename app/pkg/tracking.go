package build

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoEndTime          = errors.New("no end time for last record")
	ErrTrackingNotStarted = errors.New("start tracking first")
	ErrNoSuchProject      = errors.New("no such project")
	current               *Record
)

type Summary map[string]time.Duration

type TrackingReport struct {
	Current *Record
	Project string
	Session time.Duration
	Today   time.Duration
	Summary Summary
	Total   time.Duration
	Breaks  time.Duration
}

func init() {
	current = nil
	InitializeDatabase()
	records, err := GetAllrecords()
	if err != nil {
		log.Println("error retrieving during timetrace init ", err)
	}
	for _, record := range records {
		if record.End.IsZero() {
			current = &record
			return
		}
	}
}

func Status() *TrackingReport {
	log.Println("get tracking report")
	today, total, breaks, summary := GetTimeWorkedToday(current)
	report := TrackingReport{
		Current: current,
		Today:   today,
		Summary: summary,
		Total:   total,
		Breaks:  breaks,
	}
	log.Println(today, total, breaks, summary)
	if current != nil {
		report.Project = current.Project
		report.Session = time.Since(current.Start).Round(time.Second)
	}
	log.Println("report:", report)
	return &report
}

// func Start(p *models.Project, u *models.User) error {
func Start(name string) error {
	if current != nil {
		if err := Stop(); err != nil {
			return err
		}
	}
	project, err := GetProject(name)
	if err != nil {
		return ErrNoSuchProject
	}
	record := Record{
		ID:      uuid.New(),
		Project: project.Name,
		//User:    u.ID,
		Start: time.Now(),
	}
	if err := Saverecord(&record); err != nil {
		return err
	}
	current = &record
	return nil
}

func Stop() error {
	if current == nil {
		return ErrTrackingNotStarted
	}
	current.End = time.Now()
	if err := Saverecord(current); err != nil {
		return err
	}
	current = nil
	return nil
}

func IsToday(t time.Time) bool {
	return Equal(t, time.Now())
}

func Equal(time1, time2 time.Time) bool {
	y1, m1, d1 := time1.Date()
	y2, m2, d2 := time2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func GetTimeWorkedToday(r *Record) (time.Duration, time.Duration, time.Duration, Summary) {
	project := ""
	if r != nil {
		project = r.Project
	}
	summary := make(Summary)
	var today, total, breaks time.Duration
	records := GetTodaysRecords()
	lastStop := &time.Time{}
	for _, record := range records {
		if record.End.IsZero() {
			if record.Project == project {
				today = today + time.Since(record.Start).Round(time.Second)
			}
			total = total + time.Since(record.Start).Round(time.Second)
			summary[record.Project] = summary[record.Project] + time.Since(record.Start).Round(time.Second)
		} else {
			if record.Project == project {
				today = today + record.End.Sub(record.Start).Round(time.Second)
			}
			total = total + record.End.Sub(record.Start).Round(time.Second)
			summary[record.Project] = summary[record.Project] + record.End.Sub(record.Start).Round(time.Second)
		}
		if !lastStop.IsZero() {
			breaks = breaks + record.Start.Sub(*lastStop).Round(time.Second)
		}
		lastStop = &record.End
	}
	return today, total, breaks, summary
}

func GetTodaysRecords() []Record {
	records := []Record{}
	rows, err := GetAllrecords()
	if err != nil {
		return records
	}
	for _, row := range rows {
		if IsToday(row.Start) || IsToday(row.End) {
			records = append(records, row)
		}
	}
	return records
}

func BackupProject(name string) error {
	return nil
}

func DeleteRecordsByProject(name string) error {
	failures := false
	rows, err := GetAllrecords()
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Project == name {
			if err := DeleteRecord(row.ID.String()); err != nil {
				failures = true
				log.Println("error deleting record", err)
			}
		}
	}
	if failures {
		return errors.New("all records were not deleted")
	}
	return nil
}
