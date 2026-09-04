-- name: GetLegacyPolicyCategories :many
SELECT id, name, `order`
FROM `vatusa-old`.policy_categories
ORDER BY `order`, id;

-- name: GetLegacyPolicies :many
SELECT id, ident, category, title, slug, description, extension, effective_date, perms, visible, `order`
FROM `vatusa-old`.policies
ORDER BY category, `order`, id;

-- name: GetPolicyCategoryByTitle :one
SELECT * FROM policy_category WHERE title = ?;

-- name: GetPolicyDocumentByIdent :one
SELECT * FROM policy_document WHERE ident = ?;
