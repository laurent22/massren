package sqlkv

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

const (
	PLACEHOLDER_QUESTION_MARK = 1
	PLACEHOLDER_DOLLAR        = 2
)

type SqlKv struct {
	db              *sql.DB
	tableName       string
	driverName      string
	placeholderType int
}

type SqlKvRow struct {
	Name  string
	Value string
}

func New(db *sql.DB, tableName string) *SqlKv {
	output := new(SqlKv)
	output.db = db
	output.tableName = tableName
	output.placeholderType = PLACEHOLDER_QUESTION_MARK

	err := output.createTable()
	if err != nil {
		panic(err)
	}

	return output
}

func (this *SqlKv) SetDriverName(n string) {
	this.driverName = n
	if this.driverName == "postgres" {
		this.placeholderType = PLACEHOLDER_DOLLAR
	} else {
		this.placeholderType = PLACEHOLDER_QUESTION_MARK
	}
}

func (this *SqlKv) createTable() error {
	_, err := this.db.Exec("CREATE TABLE IF NOT EXISTS " + this.tableName + " (name TEXT NOT NULL PRIMARY KEY, value TEXT)")
	if err != nil {
		return err
	}

	// Ignore error here since there will be one if the index already exists
	this.db.Exec("CREATE INDEX name_index ON " + this.tableName + " (name)")
	return nil
}

func (this *SqlKv) placeholder(index int) string {
	if this.placeholderType == PLACEHOLDER_QUESTION_MARK {
		return "?"
	} else {
		return "$" + strconv.Itoa(index)
	}
}

func (this *SqlKv) rowByName(name string) (*SqlKvRow, error) {
	row := new(SqlKvRow)
	query := "SELECT name, value FROM " + this.tableName + " WHERE name = " + this.placeholder(1)
	err := this.db.QueryRow(query, name).Scan(&row.Name, &row.Value)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return row, nil
}

func (this *SqlKv) All() []SqlKvRow {
	rows, err := this.db.Query("SELECT name, `value` FROM " + this.tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			return []SqlKvRow{}
		} else {
			panic(err)
		}
	}

	var output []SqlKvRow
	for rows.Next() {
		var kvRow SqlKvRow
		rows.Scan(&kvRow.Name, &kvRow.Value)
		output = append(output, kvRow)
	}

	return output
}

func (this *SqlKv) String(name string) string {
	row, err := this.rowByName(name)
	if err == nil && row == nil {
		return ""
	}
	if err != nil {
		panic(err)
	}
	return row.Value
}

func (this *SqlKv) StringD(name string, defaultValue string) string {
	if !this.HasKey(name) {
		return defaultValue
	}
	return this.String(name)
}

func (this *SqlKv) SetString(name string, value string) {
	row, err := this.rowByName(name)
	var query string

	if row == nil && err == nil {
		query = "INSERT INTO " + this.tableName + " (value, name) VALUES(" + this.placeholder(1) + ", " + this.placeholder(2) + ")"
	} else {
		query = "UPDATE " + this.tableName + " SET value = " + this.placeholder(1) + " WHERE name = " + this.placeholder(2)
	}

	_, err = this.db.Exec(query, value, name)
	if err != nil {
		panic(err)
	}
}

func (this *SqlKv) Int(name string) int {
	s := this.String(name)
	if s == "" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func (this *SqlKv) IntD(name string, defaultValue int) int {
	if !this.HasKey(name) {
		return defaultValue
	}
	return this.Int(name)
}

func (this *SqlKv) SetInt(name string, value int) {
	s := strconv.Itoa(value)
	this.SetString(name, s)
}

func (this *SqlKv) Float(name string) float32 {
	s := this.String(name)
	if s == "" {
		return 0
	}

	o, err := strconv.ParseFloat(s, 32)
	if err != nil {
		panic(err)
	}
	return float32(o)
}

func (this *SqlKv) FloatD(name string, defaultValue float32) float32 {
	if !this.HasKey(name) {
		return defaultValue
	}
	return this.Float(name)
}

func (this *SqlKv) SetFloat(name string, value float32) {
	s := strconv.FormatFloat(float64(value), 'g', -1, 32)
	this.SetString(name, s)
}

func (this *SqlKv) Bool(name string) bool {
	s := this.String(name)
	return s == "1" || strings.ToLower(s) == "true"
}

func (this *SqlKv) BoolD(name string, defaultValue bool) bool {
	if !this.HasKey(name) {
		return defaultValue
	}
	return this.Bool(name)
}

func (this *SqlKv) SetBool(name string, value bool) {
	var s string
	if value {
		s = "1"
	} else {
		s = "0"
	}
	this.SetString(name, s)
}

func (this *SqlKv) Time(name string) time.Time {
	s := this.String(name)
	if s == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}

	return t
}

func (this *SqlKv) TimeD(name string, defaultValue time.Time) time.Time {
	if !this.HasKey(name) {
		return defaultValue
	}
	return this.Time(name)
}

func (this *SqlKv) SetTime(name string, value time.Time) {
	this.SetString(name, value.Format(time.RFC3339Nano))
}

func (this *SqlKv) Del(name string) {
	query := "DELETE FROM " + this.tableName + " WHERE name = " + this.placeholder(1)
	_, err := this.db.Exec(query, name)

	if err != nil {
		panic(err)
	}
}

func (this *SqlKv) Clear() {
	query := "DELETE FROM " + this.tableName
	_, err := this.db.Exec(query)

	if err != nil {
		panic(err)
	}
}

func (this *SqlKv) HasKey(name string) bool {
	row, err := this.rowByName(name)
	if row == nil && err == nil {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true
}
