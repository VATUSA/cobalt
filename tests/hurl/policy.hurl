# /policy CRUD: global-write permission gate, document_url scheme validation
# (config.IsSafeDocumentURL), the hidden-field absent-vs-explicit-false fix
# (models.PolicyDocumentRequest.Hidden as *bool), and 404/409 on
# missing/non-empty rows.

# Writes are global-write-gated (acl.ObjectPolicy); an anonymous caller is
# rejected before any session exists.
POST http://localhost:8080/policy/category
{
    "title": "Should Not Be Created",
    "sort_order": 0
}
HTTP 403

POST http://localhost:9100/mock/set-scenario
{
    "scenario": "active"
}
HTTP 204

# 900001 is granted the global "division_staff" role in seed.sql, which
# carries write:policy.
GET http://localhost:8080/login
[Options]
location: true
HTTP 200

POST http://localhost:8080/policy/category
{
    "title": "Integration Test Category",
    "sort_order": 0
}
HTTP 200
[Captures]
category_id: jsonpath "$.id"

# A javascript: (or any non-http(s)) document_url must 400 -- the staff app
# renders this value as an href, so accepting it verbatim is a stored-XSS
# primitive.
POST http://localhost:8080/policy/document
{
    "policy_category_id": {{category_id}},
    "ident": "TEST-XSS",
    "title": "Malicious Document",
    "summary": "s",
    "document_url": "javascript:alert(1)",
    "effective_date": "2026-01-01",
    "hidden": false,
    "sort_order": 0
}
HTTP 400

# A policy_category_id that doesn't exist must 400.
POST http://localhost:8080/policy/document
{
    "policy_category_id": 999999,
    "ident": "TEST-ORPHAN",
    "title": "Orphan Document",
    "summary": "s",
    "document_url": "https://example.test/doc.pdf",
    "effective_date": "2026-01-01",
    "hidden": false,
    "sort_order": 0
}
HTTP 400

# "hidden" omitted entirely on create defaults to false.
POST http://localhost:8080/policy/document
{
    "policy_category_id": {{category_id}},
    "ident": "TEST-VISIBLE",
    "title": "Visible Document",
    "summary": "s",
    "document_url": "https://example.test/visible.pdf",
    "effective_date": "2026-01-01",
    "sort_order": 0
}
HTTP 200
[Captures]
visible_document_id: jsonpath "$.id"

POST http://localhost:8080/policy/document
{
    "policy_category_id": {{category_id}},
    "ident": "TEST-HIDDEN",
    "title": "Hidden Document",
    "summary": "s",
    "document_url": "https://example.test/hidden.pdf",
    "effective_date": "2026-01-01",
    "hidden": true,
    "sort_order": 0
}
HTTP 200
[Captures]
hidden_document_id: jsonpath "$.id"

# This is the only category created against this (freshly seeded) instance
# up to this point, so it's safe to index directly rather than filter.
# GetPolicyCategories/GetAllPolicyDocuments both order by (sort_order, id),
# and both documents share sort_order 0, so the lower-id "visible" document
# (created first) sorts before "hidden".
GET http://localhost:8080/policy
HTTP 200
[Asserts]
jsonpath "$[0].id" == {{category_id}}
jsonpath "$[0].documents[0].id" == {{visible_document_id}}
jsonpath "$[0].documents[0].hidden" == false
jsonpath "$[0].documents[1].id" == {{hidden_document_id}}
jsonpath "$[0].documents[1].hidden" == true

# Updating the hidden document without an explicit "hidden" field must leave
# it hidden -- with a plain bool (rather than *bool) this used to decode to
# false and silently re-publish it.
POST http://localhost:8080/policy/document/{{hidden_document_id}}
{
    "policy_category_id": {{category_id}},
    "ident": "TEST-HIDDEN",
    "title": "Hidden Document (edited)",
    "summary": "s",
    "document_url": "https://example.test/hidden.pdf",
    "effective_date": "2026-01-01",
    "sort_order": 0
}
HTTP 200

GET http://localhost:8080/policy
HTTP 200
[Asserts]
jsonpath "$[0].documents[1].hidden" == true
jsonpath "$[0].documents[1].title" == "Hidden Document (edited)"

# A category with documents still in it must not be deletable (RESTRICT
# foreign key as of migration 000020).
DELETE http://localhost:8080/policy/category/{{category_id}}
HTTP 409

# Updating/deleting a nonexistent id is a real 404.
POST http://localhost:8080/policy/document/999999
{
    "policy_category_id": {{category_id}},
    "ident": "TEST-MISSING",
    "title": "Missing",
    "summary": "s",
    "document_url": "https://example.test/missing.pdf",
    "effective_date": "2026-01-01",
    "sort_order": 0
}
HTTP 404

DELETE http://localhost:8080/policy/document/999999
HTTP 404

DELETE http://localhost:8080/policy/document/{{visible_document_id}}
HTTP 200

DELETE http://localhost:8080/policy/document/{{hidden_document_id}}
HTTP 200

DELETE http://localhost:8080/policy/category/{{category_id}}
HTTP 200

DELETE http://localhost:8080/policy/category/{{category_id}}
HTTP 404
