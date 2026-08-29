package endpoints

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/storage"

	"github.com/labstack/echo/v5"
)

// documentFormField is the multipart field the staff app posts the chosen
// document file under.
const documentFormField = "document"

// bindPolicyDocumentRequest fills request from the incoming body and returns
// the uploaded document, if any. A multipart submission may carry a file; a
// JSON body is still accepted so a caller that already has a hosted URL keeps
// working. In both cases document_url is honoured as the existing value,
// which is what lets an edit that doesn't replace the file keep the document
// it already has.
func bindPolicyDocumentRequest(c *echo.Context, request *models.PolicyDocumentRequest) (*multipart.FileHeader, error) {
	contentType := c.Request().Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return nil, c.Bind(request)
	}

	// ParseMultipartForm will happily spill an unbounded body to temp files, so
	// cap the whole request first. The slack over MaxDocumentBytes covers the
	// multipart framing and the other form fields.
	req := c.Request()
	req.Body = http.MaxBytesReader(c.Response(), req.Body, storage.MaxDocumentBytes+(1<<20))

	// Small in-memory budget: anything bigger spills to a temp file that Go
	// removes when the request ends, rather than being held in the heap.
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		return nil, errors.New("could not read the submitted form (is the file too large?)")
	}

	form := req.MultipartForm
	field := func(name string) string {
		if values := form.Value[name]; len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}

	categoryId, _ := strconv.Atoi(field("policy_category_id"))
	sortOrder, _ := strconv.Atoi(field("sort_order"))
	hidden, _ := strconv.ParseBool(field("hidden"))

	request.PolicyCategoryId = categoryId
	request.Ident = field("ident")
	request.Title = field("title")
	request.Summary = field("summary")
	request.DocumentUrl = field("document_url")
	request.EffectiveDate = field("effective_date")
	request.Hidden = hidden
	request.SortOrder = sortOrder

	if files := form.File[documentFormField]; len(files) > 0 && files[0].Size > 0 {
		return files[0], nil
	}
	return nil, nil
}

// resolveDocumentURL uploads a submitted document and returns the URL to
// store, or falls back to the URL already on the request when no new file
// was chosen. Call it only after the permission check has passed, so an
// unauthorised request can't write to the bucket.
func resolveDocumentURL(ctx context.Context, existingURL string, document *multipart.FileHeader) (string, int, error) {
	if document == nil {
		if existingURL == "" {
			return "", http.StatusBadRequest, errors.New("a document file is required")
		}
		return existingURL, 0, nil
	}

	url, err := storage.UploadPolicyDocument(ctx, document)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidDocument) {
			return "", http.StatusBadRequest, err
		}
		return "", http.StatusInternalServerError, err
	}
	return url, 0, nil
}

func parsePolicyEffectiveDate(raw string) (time.Time, error) {
	effectiveDate, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, errors.New("effective_date must be a date in YYYY-MM-DD format")
	}
	return effectiveDate, nil
}

func GetPolicies(c *echo.Context) error {
	ctx := c.Request().Context()

	categories, err := dbconn.Queries().GetPolicyCategories(ctx)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	var documents []db.PolicyDocument
	if HasGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		documents, err = dbconn.Queries().GetAllPolicyDocuments(ctx)
	} else {
		documents, err = dbconn.Queries().GetVisiblePolicyDocuments(ctx)
	}
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	documentsByCategory := make(map[int32][]models.PolicyDocument)
	for _, document := range documents {
		documentsByCategory[document.PolicyCategoryID] = append(
			documentsByCategory[document.PolicyCategoryID], models.PolicyDocumentFromDatabase(document))
	}

	output := make([]models.PolicyCategory, len(categories))
	for i, category := range categories {
		output[i] = models.PolicyCategoryFromDatabase(category, documentsByCategory[category.ID])
	}

	return c.JSON(http.StatusOK, output)
}

func CreatePolicyCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.PolicyCategoryRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().CreatePolicyCategory(ctx, db.CreatePolicyCategoryParams{
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

func UpdatePolicyCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid category id"))
	}
	ctx := c.Request().Context()
	var request models.PolicyCategoryRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	err = dbconn.Queries().UpdatePolicyCategory(ctx, db.UpdatePolicyCategoryParams{
		Title:     request.Title,
		SortOrder: int32(request.SortOrder),
		ID:        int32(categoryId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update category"))
	}
	return GenericSuccess(c, categoryId)
}

func DeletePolicyCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid category id"))
	}
	ctx := c.Request().Context()

	err = dbconn.Queries().DeletePolicyCategory(ctx, int32(categoryId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete category"))
	}
	return GenericSuccess(c, categoryId)
}

func CreatePolicyDocument(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.PolicyDocumentRequest
	document, err := bindPolicyDocumentRequest(c, &request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	effectiveDate, err := parsePolicyEffectiveDate(request.EffectiveDate)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	// Last thing before the insert: an upload that succeeds but is then
	// followed by a rejected request leaves an orphaned object in the bucket.
	documentUrl, status, err := resolveDocumentURL(ctx, request.DocumentUrl, document)
	if err != nil {
		return GenericError(c, status, err)
	}

	now := time.Now()
	result, err := dbconn.Queries().CreatePolicyDocument(ctx, db.CreatePolicyDocumentParams{
		PolicyCategoryID: int32(request.PolicyCategoryId),
		Ident:            request.Ident,
		Title:            request.Title,
		Summary:          request.Summary,
		DocumentUrl:      documentUrl,
		EffectiveDate:    effectiveDate,
		Hidden:           request.Hidden,
		SortOrder:        int32(request.SortOrder),
		CreatedAt:        now,
		UpdatedAt:        now,
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

func UpdatePolicyDocument(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	documentId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid document id"))
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetPolicyDocumentById(ctx, int32(documentId))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("policy document not found"))
	}

	var request models.PolicyDocumentRequest
	document, err := bindPolicyDocumentRequest(c, &request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	effectiveDate, err := parsePolicyEffectiveDate(request.EffectiveDate)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	// Without a replacement file the document keeps the file it already has,
	// even if the client omitted document_url entirely.
	existingUrl := request.DocumentUrl
	if existingUrl == "" {
		existingUrl = existing.DocumentUrl
	}
	documentUrl, status, err := resolveDocumentURL(ctx, existingUrl, document)
	if err != nil {
		return GenericError(c, status, err)
	}

	err = dbconn.Queries().UpdatePolicyDocument(ctx, db.UpdatePolicyDocumentParams{
		PolicyCategoryID: int32(request.PolicyCategoryId),
		Ident:            request.Ident,
		Title:            request.Title,
		Summary:          request.Summary,
		DocumentUrl:      documentUrl,
		EffectiveDate:    effectiveDate,
		Hidden:           request.Hidden,
		SortOrder:        int32(request.SortOrder),
		UpdatedAt:        time.Now(),
		ID:               int32(documentId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update policy document"))
	}
	return GenericSuccess(c, documentId)
}

func DeletePolicyDocument(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	documentId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid document id"))
	}
	ctx := c.Request().Context()

	err = dbconn.Queries().DeletePolicyDocument(ctx, int32(documentId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete policy document"))
	}
	return GenericSuccess(c, documentId)
}
