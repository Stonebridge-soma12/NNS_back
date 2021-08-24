package message

type MessageType string

const (
	ConnectResponse           MessageType = "connect_response"
	LoginRequest              MessageType = "login_request"
	LoginResponse             MessageType = "login_response"
	JoinRequest               MessageType = "join_request"
	JoinResponse              MessageType = "join_response"
	InitDataRequest           MessageType = "init_data_request"
	InitDataResponse          MessageType = "init_data_response"
	ChangeCurrentUserResponse MessageType = "change_current_user_response"
	MoveCursorRequest         MessageType = "move_cursor_request"
	MoveCursorResponse        MessageType = "move_cursor_response"
	ExitCursorResponse        MessageType = "exit_cursor_response"
	MoveBlockRequest          MessageType = "move_block_request"
	MoveBlockResponse         MessageType = "move_block_response"
	ChangeBlockRequest        MessageType = "change_block_request"
	ChangeBlockResponse       MessageType = "change_block_request"
	CreateBlockRequest        MessageType = "create_block_request"
	CreateBlockResponse       MessageType = "create_block_response"
	ErrorResponse             MessageType = "error_response"
	LeaveRequest              MessageType = "leave_response"
	LeaveResponse             MessageType = "leave_response"
	BeforeDisconnect          MessageType = "before_disconnect"
	DisconnectResponse        MessageType = "disconnect_response"
)
