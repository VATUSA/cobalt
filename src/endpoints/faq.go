package endpoints

import (
	"errors"
	"net/http"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetFaqs(c *echo.Context) error {
	ctx := c.Request().Context()

	categories, err := dbconn.Queries().GetFaqCategories(ctx)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	items, err := dbconn.Queries().GetAllFaqItems(ctx)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	itemsByCategory := make(map[int32][]models.FaqItem)
	for _, item := range items {
		itemsByCategory[item.FaqCategoryID] = append(itemsByCategory[item.FaqCategoryID], models.FaqItemFromDatabase(item))
	}

	output := make([]models.FaqCategory, len(categories))
	for i, category := range categories {
		output[i] = models.FaqCategoryFromDatabase(category, itemsByCategory[category.ID])
	}

	return c.JSON(http.StatusOK, output)
}

func CreateFaqCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FaqCategoryRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if err := requireText("title", request.Title, 120); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().CreateFaqCategory(ctx, db.CreateFaqCategoryParams{
		Title:     request.Title,
		SortOrder: int32(request.SortOrder),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return GenericSuccess(c, int(lastInsertId))
}

func UpdateFaqCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	categoryId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FaqCategoryRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if err := requireText("title", request.Title, 120); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().UpdateFaqCategory(ctx, db.UpdateFaqCategoryParams{
		Title:     request.Title,
		SortOrder: int32(request.SortOrder),
		ID:        categoryId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update category"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("faq category not found"))
	}
	return GenericSuccess(c, int(categoryId))
}

func DeleteFaqCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	categoryId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	count, err := dbconn.Queries().CountFaqItemsInCategory(ctx, categoryId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if count > 0 {
		return GenericError(c, http.StatusConflict, errors.New("category still has items"))
	}

	result, err := dbconn.Queries().DeleteFaqCategory(ctx, categoryId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete category"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("faq category not found"))
	}
	return GenericSuccess(c, int(categoryId))
}

func CreateFaqItem(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FaqItemRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if request.FaqCategoryId <= 0 {
		return GenericError(c, http.StatusBadRequest, errors.New("faq_category_id is required"))
	}
	if err := requireText("question", request.Question, 500); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if err := requireText("answer", request.Answer, 65535); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if _, err := dbconn.Queries().GetFaqCategoryById(ctx, int32(request.FaqCategoryId)); err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("faq_category_id does not exist"))
	}

	result, err := dbconn.Queries().CreateFaqItem(ctx, db.CreateFaqItemParams{
		FaqCategoryID: int32(request.FaqCategoryId),
		Question:      request.Question,
		Answer:        request.Answer,
		SortOrder:     int32(request.SortOrder),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return GenericSuccess(c, int(lastInsertId))
}

func UpdateFaqItem(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	itemId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FaqItemRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if request.FaqCategoryId <= 0 {
		return GenericError(c, http.StatusBadRequest, errors.New("faq_category_id is required"))
	}
	if err := requireText("question", request.Question, 500); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if err := requireText("answer", request.Answer, 65535); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if _, err := dbconn.Queries().GetFaqCategoryById(ctx, int32(request.FaqCategoryId)); err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("faq_category_id does not exist"))
	}

	result, err := dbconn.Queries().UpdateFaqItem(ctx, db.UpdateFaqItemParams{
		FaqCategoryID: int32(request.FaqCategoryId),
		Question:      request.Question,
		Answer:        request.Answer,
		SortOrder:     int32(request.SortOrder),
		ID:            itemId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update item"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("faq item not found"))
	}
	return GenericSuccess(c, int(itemId))
}

func DeleteFaqItem(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	itemId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	result, err := dbconn.Queries().DeleteFaqItem(ctx, itemId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete item"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("faq item not found"))
	}
	return GenericSuccess(c, int(itemId))
}
