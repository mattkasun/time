package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// table names
	USERS_TABLE_NAME   = "users"
	PROJECT_TABLE_NAME = "projects"
	RECORDS_TABLE_NAME = "records"
	// sql verbs
	INIT_DB      = "init"
	CREATE_TABLE = "createtable"
	INSERT       = "insert"
	DELETE       = "delete"
	DELETE_ALL   = "deleteall"
	FETCH        = "fetch"
	CLOSE_DB     = "close"
	// errors
	NO_RECORDS = "no results found"
)

var ErrNoResults = errors.New("no results found")
var db *sql.DB

// SQLITE_FUNCTIONS - contains a map of the functions for sqlite
var SQLITE_FUNCTIONS = map[string]interface{}{
	INIT_DB:      sqInitDB,
	CREATE_TABLE: sqCreateTable,
	INSERT:       sqInsert,
	DELETE:       sqDeleteRecord,
	DELETE_ALL:   sqDeleteAllRecords,
	FETCH:        sqFetchRecords,
	CLOSE_DB:     sqCloseDB,
}

// Generic Functions
func getCurrentDB() map[string]interface{} {
	//config, _ := config.Get()
	//switch config.DB {
	//case "sqlite":
	return SQLITE_FUNCTIONS
	//case "postgres":
	//	return POSTGRES_FUNCTIONS
	//default:
	//	return SQLITE_FUNCTIONS
	//}
}

func InitializeDatabase() error {
	log.Println("connecting to database")
	if err := sqInitDB(); err != nil {
		//if err := getCurrentDB()[INIT_DB].(func() error)(); err != nil {
		return err
	}
	//pretty.Println(db)
	return createTables()
}

func createTables() error {
	if err := createTable(USERS_TABLE_NAME); err != nil {
		//if err := createTable(USERS_TABLE_NAME); err != nil {
		return err
	}
	if err := createTable(PROJECT_TABLE_NAME); err != nil {
		//if err := createTable(PROJECT_TABLE_NAME); err != nil {
		return err
	}
	if err := createTable(RECORDS_TABLE_NAME); err != nil {
		//if err := createTable(RECORDS_TABLE_NAME); err != nil {
		return err
	}
	return nil
}

func createTable(name string) error {
	return sqCreateTable(name)
	//return getCurrentDB()[CREATE_TABLE].(func(string) error)(name)
}

func insert(key, value, table string) error {
	return sqInsert(key, value, table)
	//return getCurrentDB()[INSERT].(func(string, string, string) error)(key, value, table)
}

func fetch(table string) (map[string]string, error) {
	return sqFetchRecords(table)
	//return getCurrentDB()[FETCH].(func(string) (map[string]string, error))(table)
}

func delete(key, table string) error {
	//return getCurrentDB()[DELETE].(func(string, string) error)(key, table)
	return sqDeleteRecord(key, table)
}

// Sqlite functions
func sqInitDB() error {
	if db != nil {
		return nil
	}
	//cfg, err := config.Get()
	//if err != nil {
	//log.Fatal("could not connect to database", err)
	//}
	//if cfg == (&config.Config{}) || cfg.DBFile == "" {
	//return errors.New("empty config file")
	//}
	//log.Println("initializing sqlite ", cfg.DBPath, cfg.DBFile)
	DBPath := "./"
	//DBPath := "/var/lib/timetrace/"
	DBFile := "timetrace.db"
	var err error

	if _, err := os.Stat(DBPath); os.IsNotExist(err) {
		if err := os.MkdirAll(DBPath, 0766); err != nil {
			log.Println("mkdir error: ", DBPath, err)
			return err
		}
	}
	path := filepath.Join(DBPath, DBFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err := os.Create(path)
		if err != nil {
			log.Println("file create error: ", err)
			return err
		}
	}
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Println("error opening sqlite database", path, err)
		return err
	}
	db.SetMaxOpenConns(1)
	return db.Ping()
	//return nil
}

func sqCreateTable(table string) error {
	query := "CREATE TABLE IF NOT EXISTS " + table + " ( key TEXT NOT NULL UNIQUE PRIMARY KEY, value TEXT)"
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		//statement, err := db.Prepare(query)
		//if err != nil {
		//	log.Println("error preparing query", err)
		//	return err
		//}
		//defer statement.Close()
		//_, err = statement.Exec()
		//if err != nil {
		log.Println("error executing statement", err)
		return err
	}
	return nil
}

func sqInsert(key, value, table string) error {
	//log.Println("sqlinsert", key, value, table)
	if key != "" && value != "" && json.Valid([]byte(value)) {
		insertSQL := "INSERT OR REPLACE INTO " + table + " (key, value) VALUES (?, ?)"
		statement, err := db.Prepare(insertSQL)
		if err != nil {
			return err
		}
		defer statement.Close()
		_, err = statement.Exec(key, value)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("invalid insert " + key + " : " + value)
}

func sqDeleteRecord(id, table string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE KEY = '%s'", table, id)
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()
	if _, err := statement.Exec(); err != nil {
		return err
	}
	log.Printf("deleted %s from %s\n", id, table)
	return nil
}

func sqDeleteAllRecords() error {
	return nil
}

func sqFetchRecords(table string) (map[string]string, error) {
	query := "SELECT * FROM " + table + " ORDER BY key"
	row, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	records := make(map[string]string)
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var key string
		var value string
		row.Scan(&key, &value)
		records[key] = value
	}
	if len(records) == 0 {
		return nil, ErrNoResults
	}
	return records, nil
}

func sqCloseDB() {
	db.Close()
}

// Record Functions
func Saverecord(record *Record) error {
	value, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return insert(record.ID.String(), string(value), RECORDS_TABLE_NAME)
}

func GetRecord(id string) (Record, error) {
	var record Record
	records, err := fetch(RECORDS_TABLE_NAME)
	if err != nil {
		return record, err
	}
	for key, value := range records {
		if key == id {
			if err := json.Unmarshal([]byte(value), &record); err != nil {
				return record, err
			}
			return record, nil
		}
	}
	return record, errors.New("no such record")
}

func GetAllrecords() ([]Record, error) {
	var records []Record
	var record Record
	rows, err := fetch(RECORDS_TABLE_NAME)
	if err != nil {
		return records, err
	}
	for _, value := range rows {
		if err := json.Unmarshal([]byte(value), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

func DeleteRecord(id string) error {
	return delete(id, RECORDS_TABLE_NAME)
}

// Project Functions
func SaveProject(p *Project) error {
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return insert(p.Name, string(value), PROJECT_TABLE_NAME)
}

func GetProject(name string) (Project, error) {
	var project Project
	records, err := fetch(PROJECT_TABLE_NAME)
	if err != nil {
		return project, err
	}
	for key, record := range records {
		if key == name {
			if err := json.Unmarshal([]byte(record), &project); err != nil {
				return project, err
			}
			return project, nil
		}
	}
	return project, errors.New("no such project")
}

func GetAllProjects() ([]Project, error) {
	var projects []Project
	var project Project
	records, err := fetch(PROJECT_TABLE_NAME)
	if err != nil {
		return projects, err
	}
	for _, record := range records {
		if err := json.Unmarshal([]byte(record), &project); err != nil {
			continue
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func DeleteProject(name string) error {
	return delete(name, PROJECT_TABLE_NAME)
}
