package model

import (
	"database/sql"
	"nns_back/util"
	"time"
)

type User struct {
	Id           int64          `db:"id"`
	Name         string         `db:"name"`
	ProfileImage sql.NullInt64  `db:"profile_image"`
	Description  sql.NullString `db:"description"`
	Email        sql.NullString `db:"email"` // TODO: e-mail verification
	WebSite      sql.NullString `db:"web_site"`
	LoginId      sql.NullString `db:"login_id"`
	LoginPw      util.NullBytes `db:"login_pw"`
	Status       util.Status    `db:"status"`
	CreateTime   time.Time      `db:"create_time"`
	UpdateTime   time.Time      `db:"update_time"`
}

func NewUser(id string, pw []byte) User {
	return User{
		Name:         "Anonymous",
		ProfileImage: sql.NullInt64{},
		Description:  sql.NullString{},
		Email:        sql.NullString{},
		WebSite:      sql.NullString{},
		LoginId: sql.NullString{
			String: id,
			Valid:  true,
		},
		LoginPw: util.NullBytes{
			Bytes: pw,
			Valid: true,
		},
		Status:     util.StatusEXIST,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}