# Query builder
- 직접 쓰려고만든 쿼리빌더
  
- SELECT, FROM, WHERE, JOIN 등 Add함수로 등록 후 사용하면 됨.
    - Add 함수 호출시 순서는 관계없음. Build 함수에서 쿼리문이 순서에 맞게 작성됨.
  
- 쿼리문 작성시 필요한 옵션을 직접 구현해서 사용 WithUserId 함수 참조.
  
- 반드시 필요한 부분을 모두 Add 후에 Apply 할 것.
### Install
```
go get github.com/elixter/Querybuilder
```

### Example
```go
type TestRepo struct {
    DB *sqlx.DB
}

func (t TestRepo) FindAll(opts ...Option) ([]User, error) {
    builder := ApplyQueryOptions(opts...)
    builder.AddSelect(defaultSelectTrainHistoryColumns).
    AddFrom(`user u`).
        AddJoin(`post p ON u.id = p.user_id`).
    
    err := builder.Build()
    if err != nil {
        return nil, err
    }
    
    rows, err := t.DB.Queryx(builder.QueryString, builder.Args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    ...
}

func WithUserId(userId int) Option {
    return OptionFunc(func (q *Query) {
        q.AddWhere("u.id = ?", user)
    })
}

func TestApplyQueryOptions(t *testing.T) {
    dbUrl := getDBInfo()
    
    db, err := sqlx.Open("mysql", dbUrl)
    if err != nil {
        t.Errorf(err.Error())
    }

    repo := TestRepo{DB: db}
    historyList, err := repo.FindAll(WithUserId(2))
    if err != nil {
        t.Errorf(err.Error())
    }

    fmt.Printf("%+v\n", historyList)
}
```