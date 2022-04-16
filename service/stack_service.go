package service

import (
	"context"
	"errors"
	"godas/model/domain"
	"godas/model/web"
	"godas/repository"

	"github.com/go-playground/validator/v10"
)

type StackService interface {
	Create(string) (web.StackResponse, error)
	FindById(string) (web.StackResponse, error)
	FindByIdFromOwner(id string, owner string) (web.StackResponse, error)
	FindAll() ([]web.StackResponse, error)
	FindAllFromOwner(string) ([]web.StackResponse, error)
	Push(string, web.ItemRequest) (web.ItemResponse, error)
	PushFromOwner(id string, owner string, request web.ItemRequest) (web.ItemResponse, error)
	Pop(string) (web.ItemResponse, error)
	PopFromOwner(id string, owner string) (web.ItemResponse, error)
}

type StackServiceImpl struct {
	stackRepository repository.StackRepository
	userRepository  repository.UserRepository
	validate        *validator.Validate
}

func NewStackService(stackRepository repository.StackRepository, userRepository repository.UserRepository, validate *validator.Validate) StackService {
	service := new(StackServiceImpl)
	service.stackRepository = stackRepository
	service.userRepository = userRepository
	service.validate = validate

	return service
}

func (service *StackServiceImpl) Create(id string) (web.StackResponse, error) {
	response := web.StackResponse{}

	user, err := service.userRepository.FindById(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	stack, err := service.stackRepository.Insert(context.Background(), domain.Stack{
		Items: []domain.Item{},
		Owner: user.ID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateData) {
			return response, ErrDuplicate
		}
		return response, err
	}

	return web.StackResponse{
		ID:    stack.ID,
		Owner: stack.Owner,
		Items: stack.Items,
	}, nil
}

func (service *StackServiceImpl) FindById(id string) (web.StackResponse, error) {
	response := web.StackResponse{}

	stack, err := service.stackRepository.FindById(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.StackResponse{
		ID:    stack.ID,
		Owner: stack.Owner,
		Items: stack.Items,
	}

	return response, nil
}

func (service *StackServiceImpl) FindByIdFromOwner(id string, owner string) (web.StackResponse, error) {
	response := web.StackResponse{}

	stacks, err := service.stackRepository.FindByOwner(context.Background(), owner)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	if len(stacks) < 1 {
		return response, ErrNotFound
	}

	for _, stack := range stacks {
		if stack.ID == id {
			return web.StackResponse{
				ID:    stack.ID,
				Owner: stack.Owner,
				Items: stack.Items,
			}, nil
		}
	}

	return response, ErrNotFound
}

func (service *StackServiceImpl) FindAll() ([]web.StackResponse, error) {
	stacks, err := service.stackRepository.FindAll(context.Background())
	if err != nil {
		return nil, err
	}

	response := []web.StackResponse{}
	for _, stack := range stacks {
		response = append(response, web.StackResponse{
			ID:    stack.ID,
			Owner: stack.Owner,
			Items: stack.Items,
		})
	}

	return response, nil
}

func (service *StackServiceImpl) FindAllFromOwner(owner string) ([]web.StackResponse, error) {
	stacks, err := service.stackRepository.FindByOwner(context.Background(), owner)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	response := []web.StackResponse{}
	for _, stack := range stacks {
		response = append(response, web.StackResponse{
			ID:    stack.ID,
			Owner: stack.Owner,
			Items: stack.Items,
		})
	}

	return response, nil
}

func (service *StackServiceImpl) Push(id string, request web.ItemRequest) (web.ItemResponse, error) {
	response := web.ItemResponse{}

	if err := service.validate.Struct(request); err != nil {
		return response, ErrBadRequest
	}

	stack, err := service.stackRepository.FindById(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}
	item := domain.Item{
		Index: uint64(len(stack.Items)),
		Name:  request.Name,
	}
	stack.Items = append(stack.Items, item)

	stack, err = service.stackRepository.Update(context.Background(), stack)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.ItemResponse{
		Index: item.Index,
		Name:  item.Name,
	}

	return response, nil
}

func (service *StackServiceImpl) PushFromOwner(id string, owner string, request web.ItemRequest) (web.ItemResponse, error) {
	response := web.ItemResponse{}

	if err := service.validate.Struct(request); err != nil {
		return response, ErrBadRequest
	}

	stacks, err := service.stackRepository.FindByOwner(context.Background(), owner)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}
	if len(stacks) < 1 {
		return response, ErrNotFound
	}

	stack := domain.Stack{}
	found := false
	for _, s := range stacks {
		if s.ID == id {
			stack = s
			found = true
			break
		}
	}
	if !found {
		return response, ErrNotFound
	}
	item := domain.Item{
		Index: uint64(len(stack.Items)),
		Name:  request.Name,
	}
	stack.Items = append(stack.Items, item)

	stack, err = service.stackRepository.Update(context.Background(), stack)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.ItemResponse{
		Index: item.Index,
		Name:  item.Name,
	}

	return response, nil
}

func (service *StackServiceImpl) Pop(id string) (web.ItemResponse, error) {
	response := web.ItemResponse{}

	stack, err := service.stackRepository.FindById(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	pos := len(stack.Items) - 1
	item := stack.Items[pos]
	stack.Items = stack.Items[0:pos]

	stack, err = service.stackRepository.Update(context.Background(), stack)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.ItemResponse{
		Index: item.Index,
		Name:  item.Name,
	}

	return response, nil
}

func (service *StackServiceImpl) PopFromOwner(id string, owner string) (web.ItemResponse, error) {
	response := web.ItemResponse{}

	stacks, err := service.stackRepository.FindByOwner(context.Background(), owner)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}
	if len(stacks) < 1 {
		return response, ErrNotFound
	}

	stack := domain.Stack{}
	found := false
	for _, s := range stacks {
		if s.ID == id {
			stack = s
			found = true
			break
		}
	}
	if !found {
		return response, ErrNotFound
	}

	pos := len(stack.Items) - 1
	item := stack.Items[pos]
	stack.Items = stack.Items[0:pos]

	stack, err = service.stackRepository.Update(context.Background(), stack)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.ItemResponse{
		Index: item.Index,
		Name:  item.Name,
	}

	return response, nil
}
