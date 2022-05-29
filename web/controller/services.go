package controller

import (
	service2 "go-search/web/service"
)

var srv *Services

type Services struct {
	Base *service2.Base
	Word *service2.Word
}

func NewServices() {
	srv = &Services{
		Base: service2.NewBase(),
		Word: service2.NewWord(),
	}
}
