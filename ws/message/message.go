package message

type MessageType string

const (
	TypeUserCreate MessageType = "create_user_response"
	TypeUserList   MessageType = "user_list_response"
	TypeUserRemove MessageType = "remove_user_request"

	TypeCursorMove MessageType = "move_cursor"

	TypeBlockCreate MessageType = "create_block"
	TypeBlockRemove MessageType = "remove_block"
	TypeBlockMove   MessageType = "move_block"
	TypeBlockChange MessageType = "change_block"

	TypeEdgeCreate MessageType = "create_edge"
	TypeEdgeRemove MessageType = "remove_edge"
)

type UserCreate struct {
	MessageType MessageType `json:"message"`
	Project     interface{} `json:"project"`
}

type User struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UserList struct {
	MessageType MessageType `json:"message"`
	Users       []User      `json:"users"`
}

type UserRemove struct {
	MessageType MessageType `json:"message"`
	User        User        `json:"user"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Cursor struct {
	User     User     `json:"user"`
	Position Position `json:"position"`
}

type CursorMove struct {
	MessageType MessageType `json:"message"`
	Cursor      Cursor      `json:"cursor"`
}

type BlockCreate struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	Block       interface{} `json:"block"`
}

type BlockRemove struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	Position    Position    `json:"position"`
}

type BlockChange struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	BlockState  interface{} `json:"blockState"`
}

type EdgeCreate struct {
	MessageType MessageType `json:"message"`
	EdgeID      string      `json:"edgeId"`
	Edge        interface{} `json:"edge"`
}

type EdgeRemove struct {
	MessageType MessageType `json:"message"`
	EdgeID      string      `json:"edgeId"`
}
