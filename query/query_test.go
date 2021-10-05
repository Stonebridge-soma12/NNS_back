package query

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"nns_back/train"
	"os"
	"testing"
)

func getDBInfo() string {
	id := os.Getenv("DBUSER")
	pw := os.Getenv("DBPW")
	ip := os.Getenv("DBIP")
	port := os.Getenv("DBPORT")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/nns?parseTime=true", id, pw, ip, port)
}

const (
	defaultSelectTrainHistoryColumns = `t.id,
								   t.user_id,
								   t.train_no,
								   t.project_id,
								   t.acc,
								   t.loss,
								   t.val_acc,
								   t.val_loss,
								   t.name,
								   t.epochs,
								   t.result_url,
								   t.status,
								   tc.id,
								   tc.train_id,
								   tc.train_dataset_url,
								   tc.valid_dataset_url,
								   tc.dataset_shuffle,
								   tc.dataset_label,
								   tc.dataset_normalization_usage,
								   tc.dataset_normalization_method,
								   tc.model_content,
								   tc.model_config,
								   tc.create_time,
								   tc.update_time`
)

func TestQuery_Apply(t *testing.T) {
	expectString := `SELECT user_id, project_no FROM train t WHERE t.id = ? ORDER BY user_id DESC LIMIT ?, ?`

	var q Query
	q.AddSelect("user_id, project_no").
		AddFrom("train t").
		AddWhere("t.id = ?", 2).
		AddLimit(0, 0).
		AddOrder("user_id DESC")

	err := q.Apply()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf(q.QueryString)

	if expectString != q.QueryString {
		t.Errorf("Result is not same")
	}
}

type TestRepo struct {
	DB *sqlx.DB
}

func (t TestRepo) FindAll(opts ...Option) ([]train.History, error) {
	query := ApplyQueryOptions(opts...)
	query.AddSelect(defaultSelectTrainHistoryColumns).
		AddFrom(`train t`).
		AddJoin(`train_config tc ON t.id = tc.train_id`).
		AddJoin(`project p ON t.project_id = p.id`)

	err := query.Apply()
	if err != nil {
		return nil, err
	}

	rows, err := t.DB.Queryx(query.QueryString, query.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historyList []train.History
	for rows.Next() {
		var history train.History
		err = rows.Scan(
			&history.Train.Id,
			&history.Train.UserId,
			&history.Train.TrainNo,
			&history.Train.ProjectId,
			&history.Train.Acc,
			&history.Train.Loss,
			&history.Train.ValAcc,
			&history.Train.ValLoss,
			&history.Train.Name,
			&history.Train.Epochs,
			&history.Train.ResultUrl,
			&history.Train.Status,
			&history.TrainConfig.Id,
			&history.TrainConfig.TrainId,
			&history.TrainConfig.TrainDatasetUrl,
			&history.TrainConfig.ValidDatasetUrl,
			&history.TrainConfig.DatasetShuffle,
			&history.TrainConfig.DatasetLabel,
			&history.TrainConfig.DatasetNormalizationUsage,
			&history.TrainConfig.DatasetNormalizationMethod,
			&history.TrainConfig.ModelContent,
			&history.TrainConfig.ModelConfig,
			&history.TrainConfig.CreateTime,
			&history.TrainConfig.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		historyList = append(historyList, history)
	}

	return historyList, nil
}

func WithTrainId(trainId int64) Option {
	return OptionFunc(func (q *Query) {
		q.AddWhere("train_id = ?", trainId)
	})
}

func TestApplyQueryOptions(t *testing.T) {
	dbUrl := getDBInfo()

	db, err := sqlx.Open("mysql", dbUrl)
	if err != nil {
		t.Errorf(err.Error())
	}

	repo := TestRepo{DB: db}
	historyList, err := repo.FindAll(WithTrainId(2))
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf("%+v\n", historyList)
}
