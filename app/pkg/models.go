package build

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uuid.UUID
	Username string
	Password string
	Admin    bool
	Updated  time.Time
}

func (a *User) IsValidPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

type Project struct {
	ID      uuid.UUID
	Name    string
	Active  bool
	Updated time.Time
}

type Record struct {
	ID      uuid.UUID
	Project string
	User    uuid.UUID
	Start   time.Time
	End     time.Time
}

func (r *Record) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

type ReportData struct {
	Project string
	Records []Record
	Sum     time.Duration
}

func ConvertToReport(records []Record) []ReportData {
	epoch, _ := time.Parse("2006-01-02", "0001-01-01")
	data := make(map[string][]Record)
	reportData := []ReportData{}
	project := ReportData{}
	for _, record := range records {
		data[record.Project] = append(data[record.Project], record)
	}
	sum := time.Duration(0)
	for k, v := range data {
		project.Project = k
		for _, item := range v {
			if item.End == epoch {
				item.End = time.Now()
			}
			project.Records = append(project.Records, item)
			timeSpent := item.End.Sub(item.Start)
			sum = sum + timeSpent
		}
		project.Sum = sum
		reportData = append(reportData, project)
	}
	return reportData
}
