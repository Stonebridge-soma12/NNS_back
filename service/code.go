package service

// response error message
type ErrMsg string

const (
	// 400
	ErrBadRequest         ErrMsg = "Bad Request"
	ErrInvalidPathParm    ErrMsg = "Invalid Path Parameter"
	ErrInvalidQueryParm   ErrMsg = "Invalid Query Parameter"
	ErrInvalidRequestBody ErrMsg = "Invalid Request Body"
	ErrInvalidFormat      ErrMsg = "Invalid Format"

	// 401
	ErrLoginRequired         ErrMsg = "Login Required"
	ErrInvalidAuthentication ErrMsg = "Invalid Authentication"

	// 404
	ErrNotFound ErrMsg = "Not Found"

	// 422
	ErrDuplicate ErrMsg = "Duplicate Entity"

	// 500
	ErrInternalServerError ErrMsg = "Internal Server Error"
)

// mysql error code
const (
	MysqlErrDupEntry uint16 = 1062
)
