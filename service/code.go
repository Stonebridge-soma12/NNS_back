package service

// response error message
type ErrMsg string

const (
	// 400
	ErrDuplicate          ErrMsg = "Duplicate Entity"
	ErrInvalidPathParm    ErrMsg = "Invalid Path Parameter"
	ErrInvalidQueryParm   ErrMsg = "Invalid Query Parameter"
	ErrInvalidRequestBody ErrMsg = "Invalid Request Body"
	ErrBadRequest         ErrMsg = "Bad Request"
	ErrNotFound           ErrMsg = "Not Found"

	// 500
	ErrInternalServerError ErrMsg = "Internal Server Error"
)

// mysql error code
const (
	MysqlErrDupEntry uint16 = 1062
)