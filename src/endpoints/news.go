package endpoints

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) CreatePost(c *echo.Context) error {
	if !h.HasGlobalPermission(c, auth.PermPostNews) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}
	ctx := c.Request().Context()
	var request models.NewsPostRequest
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = h.Queries.CreatePost(
		ctx,
		db.CreatePostParams{
			Title:     request.Title,
			Body:      request.Body,
			AuthorCid: int32(auth.GetUserCid(c)),
			PostTime:  time.Now().Unix(),
		})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "Post Created")
}

func (h EndpointHandler) DeletePost(c *echo.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid post id")
	}
	ctx := c.Request().Context()
	post, err := h.Queries.GetPostById(ctx, int32(postId))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "post not found")
	}

	if !h.HasGlobalPermission(c, auth.PermManageNews) && int(post.AuthorCid) != auth.GetUserCid(c) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}

	err = h.Queries.DeleteNewsPostById(ctx, int32(postId))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "post not found")
	}
	return c.JSON(http.StatusOK, "Post Deleted")
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
		log.Printf("Error getting recent posts: %s\n", err)
		if config.IsDevelopment() {
			return c.JSON(http.StatusInternalServerError, err)
		} else {
			return c.JSON(http.StatusInternalServerError, nil)
		}
	}

	output := models.NewsPostsFromDatabase(posts)

	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetPost(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	post, err := h.Queries.GetPostById(ctx, int32(idInt))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "post not found")
	}
	output := models.NewsPostFromDatabase(post)
	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetNewsPage(c *echo.Context) error {
	ctx := c.Request().Context()
	page := c.Param("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page")
	}
	recordsPerPage := 25
	offset := (pageInt - 1) * recordsPerPage

	news, err := h.Queries.GetNewsPostsPage(ctx, db.GetNewsPostsPageParams{
		Offset: int32(offset),
		Limit:  int32(recordsPerPage),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "error loading posts")
	}
	output := models.NewsPostsFromDatabase(news)
	return c.JSON(http.StatusOK, output)
}
