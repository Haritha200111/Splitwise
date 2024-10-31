package dto

type UserAccount struct {
	User_id      int    `json:"user_id,omitempty"`
	FirstName    string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	LastName     string `json:"lastname,omitempty" bson:"lastname,omitempty"`
	EmailId      string `json:"emailid,omitempty" bson:"emailid,omitempty"`
	UserPassword string `json:"userpassword,omitempty" bson:"userpassword,omitempty"`
}

type UserGroup struct {
	Group_id      int64    `json:"group_id,omitempty"`
	Group_name    string   `json:"group_name,omitempty"`
	Group_expense int      `json:"group_expense,omitempty"`
	Created_by    string   `json:"created_by,omitempty"`
	Users_to_add  []string `json:"users_to_add,omitempty" `
	IsSettled     bool     `json:"users_to_add,omitempty" `
}

type Split struct {
	SplitType string         `json:"splitType,omitempty"`
	GroupName string         `json:"groupname,omitempty"`
	Splitarr  map[string]int `json:"splitarr,omitempty"`
}

type AddUser struct {
	User      UserAccount `json:"user,omitempty"`
	GroupName string      `json:"groupname,omitempty"`
}

type AddUserToGroup struct {
	GroupName string `json:"groupname,omitempty"`
	EmailId   string `json:"emailid,omitempty" bson:"emailid,omitempty"`
}

type Pay struct {
	IsSettled bool   `json:"issettled,omitempty"`
	GroupName string `json:"group_name,omitempty"`
}
