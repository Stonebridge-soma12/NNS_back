package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Project struct {
	Id          int64     `db:"id"`
	UserId      int64     `db:"user_id"`
	ProjectNo   int       `db:"project_no"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Config      NullJson  `db:"config"`
	Content     NullJson  `db:"content"`
	CreateTime  time.Time `db:"create_time"`
	UpdateTime  time.Time `db:"update_time"`
}

func DefaultConfig() NullJson {
	d := `{
    "optimizer": "adam",
    "learning_rate": 0.001,
    "loss": "sparse_categorical_crossentropy",
    "metrics": ["accuracy"],
    "batch_size": 32,
    "epochs": 10
}`
	return NullJson{
		Json:  []byte(d),
		Valid: true,
	}
}

func DefaultContent() NullJson {
	d := `{
	"output": "",
	"layers": []
}`
	return NullJson{
		Json:  []byte(d),
		Valid: true,
	}
}

// NullJson represents a JSON that may be null.
// NullJson implements the Scanner interface so
// it can be used as a scan destination
type NullJson struct {
	Json  json.RawMessage
	Valid bool // Valid is true if Json is not NULL
}

// Scan implements the Scanner interface.
func (nj *NullJson) Scan(value interface{}) error {
	if value == nil {
		nj.Json, nj.Valid = nil, false
		return nil
	}
	nj.Valid = true

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprintf("failed to unmarshal Json value: %v", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	nj.Json = result
	return err
}

// Value implements the driver Valuer interface.
func (nj NullJson) Value() (driver.Value, error) {
	if !nj.Valid {
		return nil, nil
	}
	return nj.Json, nil
}

func SelectProjectCount(db *sqlx.DB, userId int64) (int, error) {
	var count int
	err := db.QueryRowx(`SELECT COUNT(*) FROM project p WHERE p.user_id = ?;`, userId).Scan(&count)

	return count, err
}

func SelectProjectList(db *sqlx.DB, userId int64, offset, limit int) ([]Project, error) {
	rows, err := db.Queryx(
		` SELECT p.id,
					   p.user_id,
					   p.project_no,
					   p.name,
					   p.description,
        			   p.config,
					   p.content,
					   p.create_time,
					   p.update_time
				FROM project p
				WHERE p.user_id = ?
				LIMIT ?, ?;`, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projectList := make([]Project, 0, limit)
	for rows.Next() {
		p := Project{}
		if err := rows.StructScan(&p); err != nil {
			return nil, err
		}

		projectList = append(projectList, p)
	}

	return projectList, nil
}

// Keyer designate where condition
type Keyer interface {
	apply(baseQuery string) (query string, args []interface{})
}

type uniqueFunc func(baseQuery string) (query string, args []interface{})

func (f uniqueFunc) apply(baseQuery string) (query string, args []interface{}) {
	return f(baseQuery)
}

func WithId(id int64) Keyer {
	return uniqueFunc(func(baseQuery string) (query string, args []interface{}) {
		return baseQuery + ` WHERE p.id = ?;`, []interface{}{id}
	})
}

func WithUserIdAndProjectNo(userId int64, projectNo int) Keyer {
	return uniqueFunc(func(baseQuery string) (query string, args []interface{}) {
		return baseQuery + ` WHERE p.user_id = ? and p.project_no = ?`, []interface{}{userId, projectNo}
	})
}

func SelectProject(db *sqlx.DB, key Keyer) (Project, error) {
	query, args := key.apply(
		`SELECT p.id,
					   p.user_id,
					   p.project_no,
					   p.name,
					   p.description,
					   p.config,
					   p.content,
					   p.create_time,
					   p.update_time
				FROM project p`)

	p := Project{}
	err := db.QueryRowx(query, args...).StructScan(&p)

	return p, err
}


func InsertProject(db *sqlx.DB, project Project) (int64, error) {
	result, err := db.NamedExec(
		`INSERT INTO project (user_id, 
                     project_no, 
                     name, 
                     description, 
                     config, 
                     content)
			VALUES (:user_id,
					:project_no,
					:name,
					:description,
					:config,
					:content);`, project)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateProject(db *sqlx.DB, project Project) error {
	_, err := db.NamedExec(
		`UPDATE project
				SET name        = :name,
					description = :description,
					config      = :config,
					content     = :content
				WHERE id = :id;`, project)

	return err
}

func DeleteProject(db sqlx.DB, userId int64, projectNo int) error {
	_, err := db.Exec(`DELETE FROM project WHERE user_id = ? AND project_no = ?;`, userId, projectNo)
	return err
}
