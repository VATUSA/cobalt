package endpoints

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) CreatePost(c *echo.Context) error {
	if !h.AssertGlobal(c, acl.ObjectNewsPost, acl.ActionWrite) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.NewsPostRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := h.Queries.CreatePost(
		ctx,
		db.CreatePostParams{
			Title:     request.Title,
			Body:      request.Body,
			AuthorCid: int32(auth.GetUserCid(c)),
			PostTime:  time.Now().Unix(),
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

func (h EndpointHandler) UpdatePost(c *echo.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid post id"))
	}
	ctx := c.Request().Context()
	var request models.NewsPostRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	post, err := h.Queries.GetPostById(ctx, int32(postId))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("post not found"))
	}

	if !h.AssertGlobal(c, acl.ObjectNewsPost, acl.ActionWrite) {
		return nil
	}
	if int(post.AuthorCid) != auth.GetUserCid(c) && h.AssertGlobal(c, acl.ObjectNewsPost, acl.ActionManageUnowned) {
		return err
	}

	err = h.Queries.UpdatePost(ctx, db.UpdatePostParams{
		Title:    request.Title,
		Body:     request.Body,
		EditTime: time.Now().Unix(),
		ID:       int32(postId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update post"))
	}
	return GenericSuccess(c, postId)
}

func (h EndpointHandler) DeletePost(c *echo.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid post id"))
	}
	ctx := c.Request().Context()
	post, err := h.Queries.GetPostById(ctx, int32(postId))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("post not found"))
	}

	if !h.AssertGlobal(c, acl.ObjectNewsPost, acl.ActionWrite) {
		return nil
	}
	if int(post.AuthorCid) != auth.GetUserCid(c) && !h.AssertGlobal(c, acl.ObjectNewsPost, acl.ActionManageUnowned) {
		return nil
	}

	err = h.Queries.DeleteNewsPostById(ctx, int32(postId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete post"))
	}
	return GenericSuccess(c, postId)
}

func (h EndpointHandler) GetLastPosts(c *echo.Context) error {
	ctx := c.Request().Context()

	count := c.Param("count")
	countInt, err := strconv.Atoi(count)
	if err != nil {
		countInt = 20
	} else if countInt < 1 {
		countInt = 1
	} else if countInt > 100 {
		countInt = 100
	}

	posts, err := h.Queries.GetRecentNewsPosts(ctx, int32(countInt))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	output := models.NewsPostsFromDatabase(posts)

	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetPost(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid post id"))
	}
	post, err := h.Queries.GetPostById(ctx, int32(idInt))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("post not found"))
	}
	output := models.NewsPostFromDatabase(post)
	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetNewsPage(c *echo.Context) error {
	ctx := c.Request().Context()
	page := c.Param("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid page"))
	}
	recordsPerPage := 25
	offset := (pageInt - 1) * recordsPerPage

	news, err := h.Queries.GetNewsPostsPage(ctx, db.GetNewsPostsPageParams{
		Offset: int32(offset),
		Limit:  int32(recordsPerPage),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("error loading posts"))
	}
	output := models.NewsPostsFromDatabase(news)
	return c.JSON(http.StatusOK, output)
}
