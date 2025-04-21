package request

import (
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type Usecase interface {
	Save(httpReq *requestEntity.HTTPRequest, httpResp *requestEntity.HTTPResponse, clientIP string) (string, error)
	GetByID(id string) (*requestEntity.RequestRecord, error)
	GetAll() ([]*requestEntity.RequestRecord, error)
	RepeatByID(id string) (string, error)
	ScanByID(id string) ([]string, []string, error)
}
