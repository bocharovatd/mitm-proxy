package request

import (
	"net/http"
)

type Handlers interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	RepeatByID(w http.ResponseWriter, r *http.Request)
}
