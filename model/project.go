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
	Status      Status    `db:"status"`
	CreateTime  time.Time `db:"create_time"`
	UpdateTime  time.Time `db:"update_time"`
}

func NewProject(userId int64, projectNo int, name, description string) Project {
	return Project{
		UserId:      userId,
		ProjectNo:   projectNo,
		Name:        name,
		Description: description,
		Config:      DefaultConfig(),
		Content:     DefaultContent(),
		Status:      EXIST,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
}

func DefaultConfig() NullJson {
	defaultValue := map[string]interface{}{
		"optimizer":     "adam",
		"learning_rate": 0.001,
		"loss":          "sparse_categorical_crossentropy",
		"metrics":       []interface{}{"accuracy"},
		"batch_size":    32,
		"epochs":        10,
	}
	defaultBytes, _ := json.Marshal(defaultValue)
	return NullJson{
		Json:  json.RawMessage(defaultBytes),
		Valid: true,
	}
}

func DefaultContent() NullJson {
	defaultValue := map[string]interface{}{
		"output": "",
		"layers": []interface{}{},
	}
	defaultBytes, _ := json.Marshal(defaultValue)
	return NullJson{
		Json:  json.RawMessage(defaultBytes),
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
	return nj.Json.MarshalJSON()
}

// SelectProjectCount if onlyExist is true, then select exist entity count
func SelectProjectCount(db *sqlx.DB, userId int64, onlyExist bool) (int, error) {
	query := `SELECT COUNT(*) FROM project p WHERE p.user_id = ?`
	if onlyExist {
		query += ` AND p.status = 'EXIST'`
	}
	query += ";"

	var count int
	err := db.QueryRowx(query, userId).Scan(&count)

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
        			   p.status,
					   p.create_time,
					   p.update_time
				FROM project p
				WHERE p.user_id = ? AND p.status = 'EXIST'
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

func SelectProject(db *sqlx.DB, userId int64, projectNo int) (Project, error) {
	p := Project{}
	err := db.QueryRowx(
		`SELECT p.id,
					   p.user_id,
					   p.project_no,
					   p.name,
					   p.description,
					   p.config,
					   p.content,
       				   p.status,
					   p.create_time,
					   p.update_time
				FROM project p
				WHERE p.user_id = ? AND p.project_no = ? AND p.status = 'EXIST';`, userId, projectNo).StructScan(&p)

	return p, err
}

func SelectProjectWithName(db *sqlx.DB, userId int64, projectName string) (Project, error) {
	p := Project{}
	err := db.QueryRowx(
		`SELECT p.id,
					   p.user_id,
					   p.project_no,
					   p.name,
					   p.description,
					   p.config,
					   p.content,
       				   p.status,
					   p.create_time,
					   p.update_time
				FROM project p
				WHERE p.user_id = ? AND p.status = 'EXIST' AND p.name = ?;`, userId, projectName).StructScan(&p)

	return p, err
}

func (p Project) Insert(db *sqlx.DB) (int64, error) {
	result, err := db.NamedExec(
		`INSERT INTO project (user_id, 
                     project_no, 
                     name, 
                     description, 
                     config, 
                     content,
                     status,
                     create_time,
                     update_time)
			VALUES (:user_id,
					:project_no,
					:name,
					:description,
					:config,
					:content,
			        :status,
			        :create_time,
			        :update_time);`, p)
	if err != nil {
		return 0, err
	}

	p.Id, err = result.LastInsertId()
	return p.Id, err
}

func (p Project) Update(db *sqlx.DB) error {
	p.UpdateTime = time.Now()

	_, err := db.NamedExec(
		`UPDATE project
				SET name        = :name,
					description = :description,
					config      = :config,
					content     = :content,
				    status 		= :status,
				    update_time = :update_time
				WHERE id = :id AND status = 'EXIST';`, p)

	return err
}

func (p Project) Delete(db *sqlx.DB) error {
	p.Status = DELETED
	return p.Update(db)
}
