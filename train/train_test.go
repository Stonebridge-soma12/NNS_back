package train

import (
	"fmt"
	"github.com/elixter/Querybuilder"
	"github.com/jmoiron/sqlx"
	"strings"
	"testing"
)

func TestTrainDbRepository_FindAll(t *testing.T) {
	dbUrl := getDBInfo()

	db, err := sqlx.Open("mysql", dbUrl)
	if err != nil {
		t.Errorf(err.Error())
	}

	var q query.Builder
	q.AddSelect(defaultSelectTrainHistoryColumns).
		AddFrom(`train t`).
		AddJoin(`train_config tc ON t.id = tc.train_id`).
		AddJoin(`project p ON t.project_id = p.id`).
		AddWhere(`p.user_id = ?`, 2).
		AddWhere(`p.project_no = ?`, 1).
		AddWhere(`t.status != ?`, "'DEL'").
		AddLimit(0, 100)

	err = q.Build()
	if err != nil {
		t.Errorf(err.Error())
	}

	var historyList []History

	rows, err := db.Queryx(q.QueryString, q.Args...)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history History
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
			t.Errorf(err.Error())
		}
		historyList = append(historyList, history)
	}

	fmt.Printf("%+v\n", historyList)
}

func TestHandler_DeleteTrainHistoryHandler(t *testing.T) {
	const expected = "UPDATE train t JOIN project p on train.project_id = project.id SET t.status = ? WHERE project.user_id = ? AND project.project_no = ? AND train.train_no = ?"

	builder := query.ApplyQueryOptions(WithProjectUserId(2), WithProjectProjectNo(1), WithTrainTrainNo(1))
	builder.AddUpdate("train t", "t.status = ?", TrainStatusDelete).
		AddJoin("project p on train.project_id = project.id")

	err := builder.Build()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Println(strings.TrimSuffix(builder.QueryString, " "))
	fmt.Println(strings.TrimSuffix(expected, " "))

	if strings.TrimSuffix(builder.QueryString, " ") != strings.TrimSuffix(expected, " ") {
		t.Errorf("Result is not same")
	}
}

func TestTrain_Update(t *testing.T) {
	const expected = "UPDATE train SET status = ?, acc = ?, loss = ?, val_acc = ?, val_loss = ?, epochs = ?, name = ?, result_url = ? WHERE id = ?"

	var train Train

	builder := query.ApplyQueryOptions()
	builder.AddUpdate(
		"train",
		"status = ?, acc = ?, loss = ?, val_acc = ?, val_loss = ?, epochs = ?, name = ?, result_url = ?",
		train.Status,
		train.Acc,
		train.Loss,
		train.ValAcc,
		train.ValLoss,
		train.Epochs,
		train.Name,
		train.ResultUrl,
	).AddWhere("id = ?", train.Id)

	err := builder.Build()
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Println(strings.TrimSuffix(builder.QueryString, " "))
	fmt.Println(strings.TrimSuffix(expected, " "))

	if strings.TrimSuffix(builder.QueryString, " ") != strings.TrimSuffix(expected, " ") {
		t.Errorf("Result is not same")
	}
}