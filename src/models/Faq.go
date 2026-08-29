package models

import "vatusa-cobalt/db"

type FaqCategoryRequest struct {
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

type FaqItemRequest struct {
	FaqCategoryId int    `json:"faq_category_id"`
	Question      string `json:"question"`
	Answer        string `json:"answer"`
	SortOrder     int    `json:"sort_order"`
}

type FaqItem struct {
	Id            int    `json:"id"`
	FaqCategoryId int    `json:"faq_category_id"`
	Question      string `json:"question"`
	Answer        string `json:"answer"`
	SortOrder     int    `json:"sort_order"`
}

type FaqCategory struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	Items     []FaqItem `json:"items"`
}

func FaqItemFromDatabase(ent db.FaqItem) FaqItem {
	return FaqItem{
		Id:            int(ent.ID),
		FaqCategoryId: int(ent.FaqCategoryID),
		Question:      ent.Question,
		Answer:        ent.Answer,
		SortOrder:     int(ent.SortOrder),
	}
}

func FaqItemsFromDatabase(ents []db.FaqItem) []FaqItem {
	items := make([]FaqItem, len(ents))
	for i, ent := range ents {
		items[i] = FaqItemFromDatabase(ent)
	}
	return items
}

func FaqCategoryFromDatabase(ent db.FaqCategory, items []FaqItem) FaqCategory {
	if items == nil {
		items = []FaqItem{}
	}
	return FaqCategory{
		Id:        int(ent.ID),
		Title:     ent.Title,
		SortOrder: int(ent.SortOrder),
		Items:     items,
	}
}
