package usecase

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bocharovatd/mitm-proxy/internal/pkg/scanner"
	"github.com/bocharovatd/mitm-proxy/internal/request"
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type RequestUsecase struct {
	requestRepository request.Repository
}

func NewRequestUsecase(requestRepo request.Repository) request.Usecase {
	return &RequestUsecase{
		requestRepository: requestRepo,
	}
}

func (usecase *RequestUsecase) Save(httpReq *requestEntity.HTTPRequest, httpResp *requestEntity.HTTPResponse, clientIP string) (string, error) {
	if httpReq == nil || httpResp == nil {
		return "", fmt.Errorf("failed to save request: request or response is nil")
	}

	id, err := usecase.requestRepository.Save(httpReq, httpResp, clientIP)
	if err != nil {
		return "", fmt.Errorf("failed to save request: %v", err)
	}
	return id, nil
}

func (usecase *RequestUsecase) GetByID(id string) (*requestEntity.RequestRecord, error) {
	record, err := usecase.requestRepository.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get request by ID %s: %v", id, err)
	}
	return record, nil
}

func (usecase *RequestUsecase) GetAll() ([]*requestEntity.RequestRecord, error) {
	records, err := usecase.requestRepository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all requests: %v", err)
	}
	return records, nil
}

func (usecase *RequestUsecase) RepeatByID(id string) (string, error) {
	originalRecord, err := usecase.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get original request: %v", err)
	}

	httpReq, err := originalRecord.Request.ToHTTPRequest()
	if err != nil {
		return "", fmt.Errorf("failed to convert to HTTP request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	newHttpReq := &requestEntity.HTTPRequest{
		Method:     originalRecord.Request.Method,
		Path:       originalRecord.Request.Path,
		GetParams:  originalRecord.Request.GetParams,
		Headers:    originalRecord.Request.Headers,
		Cookies:    originalRecord.Request.Cookies,
		PostParams: originalRecord.Request.PostParams,
		RawBody:    originalRecord.Request.RawBody,
		CreatedAt:  time.Now(),
	}

	newHttpResp := requestEntity.ParseHTTPResponse(resp, 0)

	newID, err := usecase.requestRepository.Save(newHttpReq, newHttpResp, "system")
	if err != nil {
		return "", fmt.Errorf("failed to save repeated request: %v", err)
	}

	return newID, nil
}

func (usecase *RequestUsecase) ScanByID(id string) ([]string, []string, error) {
	originalRecord, err := usecase.requestRepository.GetByID(id)
	if err != nil {
		return []string{}, []string{}, fmt.Errorf("failed to get original request: %v", err)
	}

	httpReq, err := originalRecord.Request.ToHTTPRequest()
	if err != nil {
		return []string{}, []string{}, fmt.Errorf("failed to convert to HTTP request: %v", err)
	}

	scanner := scanner.NewScanner()
	points := scanner.ScanRequest(httpReq)

	var (
		vulnerabilities []string
		scanErrors      []string
		checked         = make(map[string]bool)
	)

	for _, point := range points {
		key := point.Type + ":" + point.Name
		if checked[key] {
			continue
		}
		checked[key] = true

		isVulnerable, errMsg := scanner.TestInjection(point, httpReq)

		if errMsg != nil {
			scanErrors = append(scanErrors, fmt.Sprintf("%s '%s': %s",
				point.Type, point.Name, errMsg))
			continue
		}

		if isVulnerable {
			vulnerabilities = append(vulnerabilities,
				fmt.Sprintf("%s '%s' is vulnerable to command injection",
					point.Type, point.Name))
		}
	}

	return vulnerabilities, scanErrors, nil
}
