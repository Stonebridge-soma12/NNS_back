package message

type User struct {
	ID int `json:"id"`
}

type Connected struct {
	Type  MessageType `json:"type"`
	Users []User      `json:"users"`
}

func NewConnected(users []User) *Connected {
	return &Connected{
		Type:  MessageTypeConnected,
		Users: users,
	}
}

type UserJoined struct {
	Type MessageType `json:"type"`
	User User        `json:"user"`
}

func NewUserJoined(userID int) *UserJoined {
	return &UserJoined{
		Type: MessageTypeUserJoin,
		User: User{ID: userID},
	}
}

type UserLeft struct {
	Type   MessageType `json:"type"`
	UserID int         `json:"userId"`
}

func NewUserLeft(userID int) *UserLeft {
	return &UserLeft{
		Type:   MessageTypeUserLeft,
		UserID: userID,
	}
}