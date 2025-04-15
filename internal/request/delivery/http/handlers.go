package http

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bocharovatd/mitm-proxy/internal/request"
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type RequestHandlers struct {
	usecase request.Usecase
	tmpl    *template.Template
}

func NewRequestHandlers(requestUC request.Usecase) request.Handlers {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	return &RequestHandlers{
		usecase: requestUC,
		tmpl:    tmpl,
	}
}

func (handlers *RequestHandlers) GetAll(w http.ResponseWriter, r *http.Request) {
	records, err := handlers.usecase.GetAll()
	if err != nil {
		log.Printf("Failed to get all requests: %v", err)
		return
	}

	data := struct {
		Title   string
		Records []*requestEntity.RequestRecord
	}{
		Title:   "All Requests",
		Records: records,
	}

	if err := handlers.tmpl.ExecuteTemplate(w, "requests.html", data); err != nil {
		log.Printf("Failed to render template: %v", err)
		return
	}
}

func (handlers *RequestHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["requestID"]

	record, err := handlers.usecase.GetByID(id)
	if err != nil {
		log.Printf("Failed to get request by ID: %v", err)
		return
	}

	data := struct {
		Title  string
		Record *requestEntity.RequestRecord
	}{
		Title:  "Request Details",
		Record: record,
	}

	if err := handlers.tmpl.ExecuteTemplate(w, "request_details.html", data); err != nil {
		log.Printf("Failed to render template: %v", err)
		return
	}
}

func (handlers *RequestHandlers) RepeatByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["requestID"]

	newId, err := handlers.usecase.RepeatByID(id)
	if err != nil {
		log.Printf("Failed to repeat request: %v", err)
		return
	}

	http.Redirect(w, r, "/requests/"+newId, http.StatusSeeOther)
}
