package endpoints

import (
	"errors"
	"net/http"
	"strconv"
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
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid category id"))
	}
	ctx := c.Request().Context()
	var request models.FaqCategoryRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	err = dbconn.Queries().UpdateFaqCategory(ctx, db.UpdateFaqCategoryParams{
		Title:     request.Title,
		SortOrder: int32(request.SortOrder),
		ID:        int32(categoryId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update category"))
	}
	return GenericSuccess(c, categoryId)
}

func DeleteFaqCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid category id"))
	}
	ctx := c.Request().Context()

	err = dbconn.Queries().DeleteFaqCategory(ctx, int32(categoryId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete category"))
	}
	return GenericSuccess(c, categoryId)
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
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid item id"))
	}
	ctx := c.Request().Context()
	var request models.FaqItemRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	err = dbconn.Queries().UpdateFaqItem(ctx, db.UpdateFaqItemParams{
		FaqCategoryID: int32(request.FaqCategoryId),
		Question:      request.Question,
		Answer:        request.Answer,
		SortOrder:     int32(request.SortOrder),
		ID:            int32(itemId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update item"))
	}
	return GenericSuccess(c, itemId)
}

func DeleteFaqItem(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectFaq, acl.ActionWrite) {
		return nil
	}
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid item id"))
	}
	ctx := c.Request().Context()

	err = dbconn.Queries().DeleteFaqItem(ctx, int32(itemId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete item"))
	}
	return GenericSuccess(c, itemId)
}
