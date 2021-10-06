package repository

import (
	"github.com/Masterminds/squirrel"
	"nns_back/model"
	"nns_back/util"
)

type ProjectRepository interface {
	// SelectProjectCount if onlyExist is true, then select exist entity count
	SelectProjectCount(classifier SelectProjectClassifier, options ...SelectProjectOption) (int, error)

	SelectProjectList(classifier SelectProjectClassifier, offset, limit int, options ...SelectProjectOption) ([]model.Project, error)
	SelectProject(classifier SelectProjectClassifier, options ...SelectProjectOption) (model.Project, error)
	Insert(project model.Project) (int64, error)
	Update(project model.Project) error
	Delete(project model.Project) error
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

func ClassifiedByShareKey(key string) SelectProjectClassifier {
	return selectProjectClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"p.share_key": key})
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
	excludeProjectId int64
	status           util.Status
	sortOrder        ProjectSortOrder
	filterType       ProjectFilterType
	filterString     string
}

func newSelectProjectOption() selectProjectOption {
	return selectProjectOption{
		excludeProjectId: int64(0),
		status:           util.StatusEXIST,
		sortOrder:        OrderByCreateTimeAsc,
		filterType:       FilterByNone,
		filterString:     "",
	}
}

func (o selectProjectOption) apply(builder *squirrel.SelectBuilder) {
	// exclude project id
	switch o.excludeProjectId {
	case int64(0):
	default:
		*builder = builder.Where(squirrel.NotEq{"p.id": o.excludeProjectId})
	}

	// status
	switch o.status {
	case util.StatusNONE:
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

func WithExcludeProjectId(projectId int64) SelectProjectOption {
	return selectProjectOptionFunc(func(option *selectProjectOption) {
		option.excludeProjectId = projectId
	})
}

func WithStatus(status util.Status) SelectProjectOption {
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
