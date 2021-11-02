package message

type MessageType string

const (
	TypeUserCreate MessageType = "create_user_response"
	TypeUserList   MessageType = "user_list_response"

	TypeCursorMove MessageType = "move_cursor"

	TypeBlockCreate       MessageType = "create_block"
	TypeBlockRemove       MessageType = "remove_block"
	TypeBlockMove         MessageType = "move_block"
	TypeBlockConfigChange MessageType = "change_block_config"
	TypeBlockLabelChange  MessageType = "change_block_label"

	TypeEdgeCreate MessageType = "create_edge"
	TypeEdgeUpdate MessageType = "update_edge"

	TypeChat MessageType = "chat_msg"
)

const MessageTypeJsonTag = "message"

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UserCreate struct {
	MessageType MessageType `json:"message"`
	User        User        `json:"user"`
	Project     interface{} `json:"project"`
}

func NewUserCreate(user User, project interface{}) UserCreate {
	return UserCreate{
		MessageType: TypeUserCreate,
		User:        user,
		Project:     project,
	}
}

type UserList struct {
	MessageType MessageType `json:"message"`
	Users       []User      `json:"users"`
}

func NewUserList(users []User) UserList {
	return UserList{
		MessageType: TypeUserList,
		Users:       users,
	}
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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
}

type BlockMove struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	Position    Position    `json:"position"`
}

type BlockConfigChange struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	Config      struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"config"`
}

type BlockLabelChange struct {
	MessageType MessageType `json:"message"`
	BlockID     string      `json:"blockId"`
	Data        string      `json:"data"`
}

type EdgeCreate struct {
	MessageType MessageType   `json:"message"`
	Elements    []interface{} `json:"elements"`
}

type EdgeUpdate struct {
	MessageType MessageType   `json:"message"`
	Elements    []interface{} `json:"elements"`
}
