package models

type GenericResponse struct {
	Success bool     `json:"success"`
	Id      int      `json:"id"`
	Errors  []string `json:"errors"`
	Code    string   `json:"code,omitempty"`
}
