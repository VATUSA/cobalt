package models

import (
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

type PolicyCategoryRequest struct {
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

// PolicyDocumentRequest's Hidden and SortOrder are pointers so a create can
// tell "the caller didn't specify hidden" (defaults to false) apart from an
// update's "the caller didn't specify hidden" (keep the existing value) —
// with plain bool/int, an omitted multipart checkbox and an explicit false
// are indistinguishable, which previously let an update silently publish a
// hidden policy document.
type PolicyDocumentRequest struct {
	PolicyCategoryId int    `json:"policy_category_id"`
	Ident            string `json:"ident"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	DocumentUrl      string `json:"document_url"`
	EffectiveDate    string `json:"effective_date"`
	Hidden           *bool  `json:"hidden"`
	SortOrder        *int   `json:"sort_order"`
}

type PolicyDocument struct {
	Id               int    `json:"id"`
	PolicyCategoryId int    `json:"policy_category_id"`
	Ident            string `json:"ident"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	DocumentUrl      string `json:"document_url"`
	EffectiveDate    string `json:"effective_date"`
	Hidden           bool   `json:"hidden"`
	SortOrder        int    `json:"sort_order"`
	CreatedByCid     int    `json:"created_by_cid"`
	UpdatedByCid     int    `json:"updated_by_cid"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
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
		EffectiveDate:    ent.EffectiveDate.Format(time.DateOnly),
		Hidden:           ent.Hidden,
		SortOrder:        int(ent.SortOrder),
		CreatedByCid:     int(ent.CreatedByCid),
		UpdatedByCid:     int(ent.UpdatedByCid),
		CreatedAt:        ent.CreatedAt.Format(config.TimestampFormat),
		UpdatedAt:        ent.UpdatedAt.Format(config.TimestampFormat),
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
