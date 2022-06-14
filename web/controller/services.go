package controller

import (
	service3 "go-search/web/service"
)

var srv *Services

type Services struct {
	Base  *service3.Base
	Word  *service3.Word
	ScLog *service3.ScLog
}

func NewServices() {
	srv = &Services{
		Base:  service3.NewBase(),
		Word:  service3.NewWord(),
		ScLog: service3.NewScLog(),
	}
}
