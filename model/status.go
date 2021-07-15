package model

type Status string

const (
	StatusNONE    Status = "" // No constraints
	StatusEXIST   Status = "EXIST"
	StatusDELETED Status = "DELETED"
)
