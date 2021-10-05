package repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"nns_back/model"
	"nns_back/util"
	"time"
)

type projectMysqlRepository struct {
	db *sqlx.DB
}

func NewProjectMysqlRepository(db *sqlx.DB) ProjectRepository {
	return &projectMysqlRepository{
		db: db,
	}
}

func (r *projectMysqlRepository) SelectProjectCount(classifier SelectProjectClassifier, options ...SelectProjectOption) (int, error) {
	builder := squirrel.Select("COUNT(*)").From("project p")
	apply(&builder, classifier, options...)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build sql")
	}

	var count int
	err = r.db.QueryRowx(query, args...).Scan(&count)

	return count, err
}

func (r *projectMysqlRepository) SelectProjectList(classifier SelectProjectClassifier, offset, limit int, options ...SelectProjectOption) ([]model.Project, error) {
	builder := squirrel.
		Select("p.id",
			"p.share_key",
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

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projectList := make([]model.Project, 0, limit)
	for rows.Next() {
		p := model.Project{}
		if err := rows.StructScan(&p); err != nil {
			return nil, err
		}

		projectList = append(projectList, p)
	}

	return projectList, nil
}

func (r *projectMysqlRepository) SelectProject(classifier SelectProjectClassifier, options ...SelectProjectOption) (model.Project, error) {
	builder := squirrel.
		Select("p.id",
			"p.share_key",
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
		return model.Project{}, errors.Wrap(err, "failed to build sql")
	}

	p := model.Project{}
	err = r.db.QueryRowx(query, args...).StructScan(&p)

	return p, err
}

func (r *projectMysqlRepository) Insert(project model.Project) (int64, error) {
	result, err := r.db.NamedExec(
		`INSERT INTO project (
                     share_key,
                     user_id, 
                     project_no, 
                     name, 
                     description, 
                     config, 
                     content,
                     status,
                     create_time,
                     update_time)
			VALUES (:share_key,
			        :user_id,
					:project_no,
					:name,
					:description,
					:config,
					:content,
			        :status,
			        :create_time,
			        :update_time);`, project)
	if err != nil {
		return 0, err
	}

	project.Id, err = result.LastInsertId()
	return project.Id, err
}

func (r *projectMysqlRepository) Update(project model.Project) error {
	project.UpdateTime = time.Now()

	_, err := r.db.NamedExec(
		`UPDATE project
				SET share_key	= :share_key,
				    name        = :name,
					description = :description,
					config      = :config,
					content     = :content,
				    status 		= :status,
				    update_time = :update_time
				WHERE id = :id AND status = 'EXIST';`, project)

	return err
}

func (r *projectMysqlRepository) Delete(project model.Project) error {
	project.Status = util.StatusDELETED
	return r.Update(project)
}
