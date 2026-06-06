package handlers

import "user_service/src/repository"

type Handler struct {
	repository repository.UserRepository
}

func NewHandler(repository repository.UserRepository) Handler {
	return Handler{repository: repository}
}
