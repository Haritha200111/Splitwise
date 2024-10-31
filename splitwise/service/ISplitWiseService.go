package service

import (
	"context"
	"split/splitwise/dto"
)

type ISplitWiseService interface {
	AddUser(request *dto.UserAccount) (*dto.Response, error)
	CreateGroup(request *dto.Request) (*dto.Response, error)
	AddUserToGroup(ctx context.Context, request *dto.AddUserToGroup) (*dto.Response, error)
	DeleteGroup(ctx context.Context, request *dto.UserGroup) (*dto.Response, error)
	Payment(ctx context.Context, request *dto.Pay) (*dto.Response, error)
	Split(request *dto.Split) (*dto.Response, error)
}
