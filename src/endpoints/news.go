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
