package models

import (
	"time"
	"vatusa-cobalt/db"
)

type PolicyCategoryRequest struct {
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

type PolicyDocumentRequest struct {
	PolicyCategoryId int    `json:"policy_category_id"`
	Ident            string `json:"ident"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	DocumentUrl      string `json:"document_url"`
	EffectiveDate    string `json:"effective_date"`
	Hidden           bool   `json:"hidden"`
	SortOrder        int    `json:"sort_order"`
}

type PolicyDocument struct {
	Id                int       `json:"id"`
	PolicyCategoryId  int       `json:"policy_category_id"`
	Ident             string    `json:"ident"`
	Title             string    `json:"title"`
	Summary           string    `json:"summary"`
	DocumentUrl       string    `json:"document_url"`
	EffectiveDate     time.Time `json:"effective_date"`
	Hidden            bool      `json:"hidden"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PolicyCategory struct {
	Id        int              `json:"id"`
	Title     string           `json:"title"`
	SortOrder int              `json:"sort_order"`
	Documents []PolicyDocument `json:"documents"`
}

func PolicyDocumentFromDatabase(ent db.PolicyDocument) PolicyDocument {
	return PolicyDocument{
		Id:               int(ent.ID),
		PolicyCategoryId: int(ent.PolicyCategoryID),
		Ident:            ent.Ident,
		Title:            ent.Title,
		Summary:          ent.Summary,
		DocumentUrl:      ent.DocumentUrl,
		EffectiveDate:    ent.EffectiveDate,
		Hidden:           ent.Hidden,
		SortOrder:        int(ent.SortOrder),
		CreatedAt:        ent.CreatedAt,
		UpdatedAt:        ent.UpdatedAt,
	}
}

func PolicyDocumentsFromDatabase(ents []db.PolicyDocument) []PolicyDocument {
	documents := make([]PolicyDocument, len(ents))
	for i, ent := range ents {
		documents[i] = PolicyDocumentFromDatabase(ent)
	}
	return documents
}

func PolicyCategoryFromDatabase(ent db.PolicyCategory, documents []PolicyDocument) PolicyCategory {
	if documents == nil {
		documents = []PolicyDocument{}
	}
	return PolicyCategory{
		Id:        int(ent.ID),
		Title:     ent.Title,
		SortOrder: int(ent.SortOrder),
		Documents: documents,
	}
}
