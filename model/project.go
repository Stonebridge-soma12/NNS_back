package model

import (
	"encoding/json"
	"github.com/Masterminds/squirrel"
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
		Status:      StatusEXIST,
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

// SelectProjectClassifier is conditions for classifying a project
type SelectProjectClassifier interface {
	classify(builder *squirrel.SelectBuilder)
}

type selectProjectClassifierFunc func(builder *squirrel.SelectBuilder)

func (f selectProjectClassifierFunc) classify(builder *squirrel.SelectBuilder) {
	f(builder)
}

func ClassifiedByUserId(userId int64) SelectProjectClassifier {
	return selectProjectClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"p.user_id": userId})
	})
}

func ClassifiedByProjectNo(userId int64, projectNo int) SelectProjectClassifier {
	return selectProjectClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"p.user_id": userId, "p.project_no": projectNo})
	})
}

func ClassifiedByProjectName(userId int64, projectName string) SelectProjectClassifier {
	return selectProjectClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"p.user_id": userId, "p.name": projectName})
	})
}

// ProjectSortOrder is define sort order
type ProjectSortOrder int

const (
	OrderByCreateTimeAsc ProjectSortOrder = iota
	OrderByCreateTimeDesc
	OrderByUpdateTimeAsc
	OrderByUpdateTimeDesc
)

// ProjectFilterType is define filter types
type ProjectFilterType int

const (
	FilterByNone ProjectFilterType = iota
	FilterByName
	FilterByNameLike
	FilterByDescription
	FilterByDescriptionLike
	FilterByNameOrDescription
	FilterByNameOrDescriptionLike
)

// SelectProjectOption is optional conditions for classifying a project
type SelectProjectOption interface {
	apply(option *selectProjectOption)
}

type selectProjectOption struct {
	status       Status
	sortOrder    ProjectSortOrder
	filterType   ProjectFilterType
	filterString string
}

func newSelectProjectOption() selectProjectOption {
	return selectProjectOption{
		status:       StatusEXIST,
		sortOrder:    OrderByCreateTimeAsc,
		filterType:   FilterByNone,
		filterString: "",
	}
}

func (o selectProjectOption) apply(builder *squirrel.SelectBuilder) {
	// status
	switch o.status {
	case StatusNONE:
	default:
		*builder = builder.Where(squirrel.Eq{"p.status": o.status})
	}

	// order by
	switch o.sortOrder {
	case OrderByCreateTimeAsc:
		*builder = builder.OrderBy("p.id ASC")
	case OrderByCreateTimeDesc:
		*builder = builder.OrderBy("p.id DESC")
	case OrderByUpdateTimeAsc:
		*builder = builder.OrderBy("p.update_time ASC")
	case OrderByUpdateTimeDesc:
		*builder = builder.OrderBy("p.update_time DESC")
	}

	// filter
	switch o.filterType {
	case FilterByNone:
	case FilterByName:
		*builder = builder.Where(squirrel.Eq{"p.name": o.filterString})
	case FilterByNameLike:
		*builder = builder.Where(squirrel.Like{"p.name": "%" + o.filterString + "%"})
	case FilterByDescription:
		*builder = builder.Where(squirrel.Eq{"p.description": o.filterString})
	case FilterByDescriptionLike:
		*builder = builder.Where(squirrel.Like{"p.description": "%" + o.filterString + "%"})
	case FilterByNameOrDescription:
		*builder = builder.Where(squirrel.Or{squirrel.Eq{"p.name": o.filterString}, squirrel.Eq{"p.description": o.filterString}})
	case FilterByNameOrDescriptionLike:
		*builder = builder.Where(squirrel.Or{squirrel.Like{"p.name": "%" + o.filterString + "%"}, squirrel.Like{"p.description": "%" + o.filterString + "%"}})
	}
}

type selectProjectOptionFunc func(option *selectProjectOption)

func (f selectProjectOptionFunc) apply(option *selectProjectOption) {
	f(option)
}

func WithStatus(status Status) SelectProjectOption {
	return selectProjectOptionFunc(func(option *selectProjectOption) {
		option.status = status
	})
}

func OrderBy(order ProjectSortOrder) SelectProjectOption {
	return selectProjectOptionFunc(func(option *selectProjectOption) {
		option.sortOrder = order
	})
}

func WithFilter(filterType ProjectFilterType, filterString string) SelectProjectOption {
	return selectProjectOptionFunc(func(option *selectProjectOption) {
		option.filterType = filterType
		option.filterString = filterString
	})
}

func apply(builder *squirrel.SelectBuilder, classifier SelectProjectClassifier, options ...SelectProjectOption) {
	classifier.classify(builder)
	option := newSelectProjectOption()
	for _, opt := range options {
		opt.apply(&option)
	}
	option.apply(builder)
}

// SelectProjectCount if onlyExist is true, then select exist entity count
func SelectProjectCount(db *sqlx.DB, classifier SelectProjectClassifier, options ...SelectProjectOption) (int, error) {
	builder := squirrel.Select("COUNT(*)").From("project p")
	apply(&builder, classifier, options...)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build sql")
	}

	var count int
	err = db.QueryRowx(query, args...).Scan(&count)

	return count, err
}

func SelectProjectList(db *sqlx.DB, classifier SelectProjectClassifier, offset, limit int, options ...SelectProjectOption) ([]Project, error) {
	builder := squirrel.
		Select("p.id",
			"p.user_id",
			"p.project_no",
			"p.name",
			"p.description",
			"p.config",
			"p.content",
			"p.status",
			"p.create_time",
			"p.update_time").
		From("project p").
		Offset(uint64(offset)).
		Limit(uint64(limit))
	apply(&builder, classifier, options...)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql")
	}

	rows, err := db.Queryx(query, args...)
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

func SelectProject(db *sqlx.DB, classifier SelectProjectClassifier, options ...SelectProjectOption) (Project, error) {
	builder := squirrel.
		Select("p.id",
			"p.user_id",
			"p.project_no",
			"p.name",
			"p.description",
			"p.config",
			"p.content",
			"p.status",
			"p.create_time",
			"p.update_time").
		From("project p")
	apply(&builder, classifier, options...)
	query, args, err := builder.ToSql()
	if err != nil {
		return Project{}, errors.Wrap(err, "failed to build sql")
	}

	p := Project{}
	err = db.QueryRowx(query, args...).StructScan(&p)

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
	p.Status = StatusDELETED
	return p.Update(db)
}
