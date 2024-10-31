package dto

type Request struct {
	User_id       int            `json:"user_id,omitempty"`
	FirstName     string         `json:"firstname,omitempty"`
	LastName      string         `json:"lastname,omitempty"`
	EmailId       string         `json:"emailid,omitempty"`
	UserPassword  string         `json:"userpassword,omitempty"`
	Group_id      int64          `json:"group_id,omitempty"`
	Group_name    string         `json:"group_name,omitempty"`
	Group_expense int            `json:"group_expense,omitempty"`
	Created_by    string         `json:"created_by,omitempty"`
	Users_to_add  []string       `json:"users_to_add,omitempty" `
	IsSettled     bool           `json:"issetteled,omitempty" `
	SplitType     string         `json:"splitType,omitempty"`
	GroupName     string         `json:"groupname,omitempty"`
	Splitarr      map[string]int `json:"splitarr,omitempty"`
	User          UserAccount    `json:"user,omitempty"`
}
