package request

import (
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type Repository interface {
	Save(req *requestEntity.HTTPRequest, resp *requestEntity.HTTPResponse, clientIP string) (string, error)
	GetByID(id string) (*requestEntity.RequestRecord, error)
	GetAll() ([]*requestEntity.RequestRecord, error)
}
