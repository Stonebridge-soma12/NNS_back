# Query builder
- 직접 쓰려고만든 쿼리빌더
- SELECT, FROM, WHERE, JOIN 등 Add함수로 등록 후 사용하면 됨.
- Option 기능도 직접 구현해서 사용. (아래 WithTrainId)참조
- 반드시 필요한 부분을 모두 Add 후에 Apply 할 것.
### Example
```go
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
```