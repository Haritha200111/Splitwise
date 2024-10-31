package service

import (
	"split/splitwise/dto"
)

type ISplitWiseService interface {
	AddUser(request *dto.UserAccount) (*dto.Response, error)
	CreateGroup(request *dto.UserGroup) (*dto.Response, error)
}
