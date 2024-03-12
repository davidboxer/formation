package core

import (
	"github.com/davidboxer/formation/resources/common"
	v1 "k8s.io/api/core/v1"
)

type Service struct {
	*common.SimpleResource[*v1.Service]
}

func NewService(service *v1.Service) *Service {
	return &Service{
		SimpleResource: common.NewSimpleResource("service", service),
	}
}
