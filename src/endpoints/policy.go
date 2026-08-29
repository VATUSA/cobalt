package endpoints

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/storage"

	"github.com/labstack/echo/v5"
)

// documentFormField is the multipart field the staff app posts the chosen
// document file under.
const documentFormField = "document"

// errDocumentRequired marks resolveDocumentURL's "no file and no existing
// URL" case so callers can distinguish it from a storage failure via
// errors.Is instead of a magic status-code return.
var errDocumentRequired = errors.New("a document file is required")

// bindPolicyDocumentRequest fills request from the incoming body and returns
// the uploaded document, if any. A multipart submission may carry a file; a
// JSON body is still accepted so a caller that already has a hosted URL keeps
// working. In both cases document_url is honoured as the existing value,
// which is what lets an edit that doesn't replace the file keep the document
// it already has.
func bindPolicyDocumentRequest(c *echo.Context, request *models.PolicyDocumentRequest) (*multipart.FileHeader, error) {
	if !isMultipartRequest(c) {
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

	categoryId, err := strconv.Atoi(field("policy_category_id"))
	if err != nil {
		return nil, errors.New("policy_category_id must be a number")
	}
	request.PolicyCategoryId = categoryId
	request.Ident = field("ident")
	request.Title = field("title")
	request.Summary = field("summary")
	request.DocumentUrl = field("document_url")
	request.EffectiveDate = field("effective_date")

	if raw := field("hidden"); raw != "" {
		hidden, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, errors.New("hidden must be a boolean")
		}
		request.Hidden = &hidden
	}
	if raw := field("sort_order"); raw != "" {
		sortOrder, err := strconv.Atoi(raw)
		if err != nil {
			return nil, errors.New("sort_order must be a number")
		}
		request.SortOrder = &sortOrder
	}

	if files := form.File[documentFormField]; len(files) > 0 {
		if files[0].Size == 0 {
			return nil, errors.New("uploaded document is empty")
		}
		return files[0], nil
	}
	return nil, nil
}

// resolveDocumentURL uploads a submitted document and returns the URL to
// store, or falls back to the URL already on the request when no new file
// was chosen. Call it only after the permission check has passed, so an
// unauthorised request can't write to the bucket.
func resolveDocumentURL(ctx context.Context, existingURL string, document *multipart.FileHeader) (string, error) {
	if document == nil {
		if existingURL == "" {
			return "", errDocumentRequired
		}
		return existingURL, nil
	}
	return storage.UploadPolicyDocument(ctx, document)
}

func parsePolicyEffectiveDate(raw string) (time.Time, error) {
	effectiveDate, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, errors.New("effective_date must be a date in YYYY-MM-DD format")
	}
	return effectiveDate, nil
}

// documentURLStatus maps a resolveDocumentURL/document-validation error to
// the HTTP status a handler should answer with.
func documentURLStatus(err error) int {
	if errors.Is(err, errDocumentRequired) || errors.Is(err, storage.ErrInvalidDocument) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
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
	if err := requireText("title", request.Title, 120); err != nil {
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
	categoryId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()
	var request models.PolicyCategoryRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if err := requireText("title", request.Title, 120); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().UpdatePolicyCategory(ctx, db.UpdatePolicyCategoryParams{
		Title:     request.Title,
		SortOrder: int32(request.SortOrder),
		ID:        categoryId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update category"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("policy category not found"))
	}
	return GenericSuccess(c, int(categoryId))
}

func DeletePolicyCategory(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	categoryId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	count, err := dbconn.Queries().CountPolicyDocumentsInCategory(ctx, categoryId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if count > 0 {
		return GenericError(c, http.StatusConflict, errors.New("category still has documents"))
	}

	result, err := dbconn.Queries().DeletePolicyCategory(ctx, categoryId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete category"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("policy category not found"))
	}
	return GenericSuccess(c, int(categoryId))
}

// validatePolicyDocumentFields checks everything except document_url, which
// isn't known until after resolveDocumentURL runs. Called before the upload
// so a malformed ident/title/summary/category doesn't orphan an uploaded
// file in the bucket.
func validatePolicyDocumentFields(request *models.PolicyDocumentRequest) error {
	if request.PolicyCategoryId <= 0 {
		return errors.New("policy_category_id is required")
	}
	if err := requireText("ident", request.Ident, 20); err != nil {
		return err
	}
	if err := requireText("title", request.Title, 255); err != nil {
		return err
	}
	if len(request.Summary) > 500 {
		return errors.New("summary must be 500 characters or fewer")
	}
	return nil
}

// validateDocumentURL checks the resolved document_url — the caller-supplied
// URL on the JSON path, or the freshly uploaded object's URL on the
// multipart path. Rejects a javascript: (or other non-http(s)) scheme, since
// the staff app renders this value as an href.
func validateDocumentURL(documentUrl string) error {
	if err := requireText("document_url", documentUrl, 500); err != nil {
		return err
	}
	if !config.IsSafeDocumentURL(documentUrl) {
		return fmt.Errorf("document_url must be an https URL")
	}
	return nil
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
	if err := validatePolicyDocumentFields(&request); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if _, err := dbconn.Queries().GetPolicyCategoryById(ctx, int32(request.PolicyCategoryId)); err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("policy_category_id does not exist"))
	}

	// Last thing before the insert: an upload that succeeds but is then
	// followed by a rejected request leaves an orphaned object in the bucket.
	documentUrl, err := resolveDocumentURL(ctx, request.DocumentUrl, document)
	if err != nil {
		return GenericError(c, documentURLStatus(err), err)
	}
	if err := validateDocumentURL(documentUrl); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	hidden := false
	if request.Hidden != nil {
		hidden = *request.Hidden
	}
	sortOrder := 0
	if request.SortOrder != nil {
		sortOrder = *request.SortOrder
	}

	userCid := int32(auth.GetUserCid(c))
	now := time.Now()
	result, err := dbconn.Queries().CreatePolicyDocument(ctx, db.CreatePolicyDocumentParams{
		PolicyCategoryID: int32(request.PolicyCategoryId),
		Ident:            request.Ident,
		Title:            request.Title,
		Summary:          request.Summary,
		DocumentUrl:      documentUrl,
		EffectiveDate:    effectiveDate,
		Hidden:           hidden,
		SortOrder:        int32(sortOrder),
		CreatedByCid:     userCid,
		UpdatedByCid:     userCid,
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
	documentId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetPolicyDocumentById(ctx, documentId)
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

	if err := validatePolicyDocumentFields(&request); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if _, err := dbconn.Queries().GetPolicyCategoryById(ctx, int32(request.PolicyCategoryId)); err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("policy_category_id does not exist"))
	}

	// Without a replacement file the document keeps the file it already has,
	// even if the client omitted document_url entirely.
	existingUrl := request.DocumentUrl
	if existingUrl == "" {
		existingUrl = existing.DocumentUrl
	}
	documentUrl, err := resolveDocumentURL(ctx, existingUrl, document)
	if err != nil {
		return GenericError(c, documentURLStatus(err), err)
	}
	if err := validateDocumentURL(documentUrl); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	// Absent from the request means "keep the existing value", which is what
	// stops an update that omits hidden/sort_order from resetting them.
	hidden := existing.Hidden
	if request.Hidden != nil {
		hidden = *request.Hidden
	}
	sortOrder := int(existing.SortOrder)
	if request.SortOrder != nil {
		sortOrder = *request.SortOrder
	}

	result, err := dbconn.Queries().UpdatePolicyDocument(ctx, db.UpdatePolicyDocumentParams{
		PolicyCategoryID: int32(request.PolicyCategoryId),
		Ident:            request.Ident,
		Title:            request.Title,
		Summary:          request.Summary,
		DocumentUrl:      documentUrl,
		EffectiveDate:    effectiveDate,
		Hidden:           hidden,
		SortOrder:        int32(sortOrder),
		UpdatedByCid:     int32(auth.GetUserCid(c)),
		UpdatedAt:        time.Now(),
		ID:               documentId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update policy document"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("policy document not found"))
	}
	return GenericSuccess(c, int(documentId))
}

func DeletePolicyDocument(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectPolicy, acl.ActionWrite) {
		return nil
	}
	documentId, ok := parseId32(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	result, err := dbconn.Queries().DeletePolicyDocument(ctx, documentId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete policy document"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("policy document not found"))
	}
	return GenericSuccess(c, int(documentId))
}
