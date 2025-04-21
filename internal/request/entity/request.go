package entity

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HTTPRequest struct {
	Method     string                 `bson:"method"`
	Path       string                 `bson:"path"`
	GetParams  map[string]interface{} `bson:"get_params"`
	Headers    map[string]string      `bson:"headers"`
	Cookies    map[string]string      `bson:"cookies"`
	PostParams map[string]interface{} `bson:"post_params"`
	RawBody    string                 `bson:"raw_body"`
	CreatedAt  time.Time              `bson:"created_at"`
}

type HTTPResponse struct {
	Code     int               `bson:"code"`
	Message  string            `bson:"message"`
	Headers  map[string]string `bson:"headers"`
	Body     string            `bson:"body"`
	Duration time.Duration     `bson:"duration"`
}

type RequestRecord struct {
	ID       primitive.ObjectID `bson:"_id"`
	Request  HTTPRequest        `bson:"request"`
	Response HTTPResponse       `bson:"response"`
	Metadata struct {
		Timestamp time.Time `bson:"timestamp"`
		ClientIP  string    `bson:"client_ip"`
	} `bson:"metadata"`
}

func (r *HTTPRequest) ToHTTPRequest() (*http.Request, error) {
	// Собираем URL с параметрами
	u := &url.URL{
		Scheme: "https",
		Path:   r.Path,
		Host:   r.Headers["Host"], // Добавляем Host в URL
	}

	query := u.Query()
	for k, v := range r.GetParams {
		if s, ok := v.(string); ok {
			query.Set(k, s)
		}
	}
	u.RawQuery = query.Encode()

	// Создаем базовый запрос
	req, err := http.NewRequest(r.Method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки (кроме Host, который уже обработан)
	for k, v := range r.Headers {
		if k != "Host" { // Исключаем Host из заголовков
			req.Header.Set(k, v)
		}
	}

	// Устанавливаем Host отдельно
	if host, exists := r.Headers["Host"]; exists {
		req.Host = host // Устанавливаем Host для запроса
	}

	// Устанавливаем cookies
	for k, v := range r.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	// Обрабатываем тело запроса
	if r.RawBody != "" {
		req.Body = io.NopCloser(strings.NewReader(r.RawBody))
		req.ContentLength = int64(len(r.RawBody))
	} else if len(r.PostParams) > 0 {
		form := url.Values{}
		for k, v := range r.PostParams {
			if s, ok := v.(string); ok {
				form.Set(k, s)
			}
		}
		encoded := form.Encode()
		req.Body = io.NopCloser(strings.NewReader(encoded))
		req.ContentLength = int64(len(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}

func ParseHTTPRequest(req *http.Request) *HTTPRequest {
	// Парсинг URL и параметров
	queryParams := parseQuery(req.URL.RawQuery)

	// Парсинг cookies
	cookies := parseCookies(req.Header.Get("Cookie"))

	// Парсинг тела запроса
	var postParams map[string]interface{}
	var rawBody string
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело
		rawBody = string(bodyBytes)

		if req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			postParams = parseQuery(string(bodyBytes))
		}
	}

	headers := flattenHeaders(req.Header)

	// Добавляем Host, если он есть в запросе
	if req.Host != "" {
		headers["Host"] = req.Host
	}

	return &HTTPRequest{
		Method:     req.Method,
		Path:       req.URL.Path,
		GetParams:  queryParams,
		Headers:    headers,
		Cookies:    cookies,
		PostParams: postParams,
		RawBody:    rawBody,
		CreatedAt:  time.Now(),
	}
}

func ParseHTTPResponse(resp *http.Response, duration time.Duration) *HTTPResponse {
	var reader io.Reader = resp.Body

	// Обработка gzip
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzReader, err := gzip.NewReader(resp.Body)
		if err == nil {
			reader = gzReader
			defer gzReader.Close()
		}
	}

	bodyBytes, _ := io.ReadAll(reader)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело

	return &HTTPResponse{
		Code:     resp.StatusCode,
		Message:  resp.Status,
		Headers:  flattenHeaders(resp.Header),
		Body:     string(bodyBytes),
		Duration: duration,
	}
}

func parseQuery(query string) map[string]interface{} {
	values, _ := url.ParseQuery(query)
	result := make(map[string]interface{})
	for k, v := range values {
		if len(v) == 1 {
			result[k] = v[0]
		} else {
			result[k] = v
		}
	}
	return result
}

func parseCookies(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	for _, c := range strings.Split(cookieHeader, ";") {
		parts := strings.SplitN(strings.TrimSpace(c), "=", 2)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}
	return cookies
}

func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
