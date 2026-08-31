# /faq CRUD: global-write permission gate, 400 on a malformed/nonexistent
# faq_category_id (endpoints/faq.go), out-of-int32-range path ids rejected
# rather than silently truncated (endpoints/params.go parseId32), and 404/409
# on operations against missing/non-empty rows.

# Writes are global-write-gated (acl.ObjectFaq); an anonymous caller is
# rejected before any session exists.
POST http://localhost:8080/faq/category
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
# carries write:faq.
GET http://localhost:8080/login
[Options]
location: true
HTTP 200

POST http://localhost:8080/faq/category
{
    "title": "Integration Test Category",
    "sort_order": 0
}
HTTP 200
[Captures]
category_id: jsonpath "$.id"

# An id outside int32 range must 400 -- strconv.Atoi + int32(...) used to
# truncate 4294967297 down to 1 and silently update whatever category that
# was.
POST http://localhost:8080/faq/category/4294967297
{
    "title": "Should Not Apply",
    "sort_order": 0
}
HTTP 400

# A faq_category_id that doesn't exist must 400, not insert a row that's
# permanently unreachable through GET /faq (which groups by category).
POST http://localhost:8080/faq/item
{
    "faq_category_id": 999999,
    "question": "Does this orphan?",
    "answer": "No.",
    "sort_order": 0
}
HTTP 400

POST http://localhost:8080/faq/item
{
    "faq_category_id": {{category_id}},
    "question": "Is this a test question?",
    "answer": "Yes.",
    "sort_order": 0
}
HTTP 200
[Captures]
item_id: jsonpath "$.id"

POST http://localhost:8080/faq/item/{{item_id}}
{
    "faq_category_id": {{category_id}},
    "question": "Is this an updated question?",
    "answer": "Still yes.",
    "sort_order": 1
}
HTTP 200

# This is the only category created against this (freshly seeded) instance
# up to this point, so it's safe to index directly rather than filter.
GET http://localhost:8080/faq
HTTP 200
[Asserts]
jsonpath "$[0].id" == {{category_id}}
jsonpath "$[0].items[0].question" == "Is this an updated question?"

# A category with items still in it must not be deletable (faq_item has a
# RESTRICT foreign key onto faq_category as of migration 000018).
DELETE http://localhost:8080/faq/category/{{category_id}}
HTTP 409

# Updating/deleting a nonexistent id is a real 404 (RowsAffected == 0), not
# the false {"success":true} the original :exec-based queries returned.
POST http://localhost:8080/faq/item/999999
{
    "faq_category_id": {{category_id}},
    "question": "Does not exist",
    "answer": "N/A",
    "sort_order": 0
}
HTTP 404

DELETE http://localhost:8080/faq/item/999999
HTTP 404

DELETE http://localhost:8080/faq/item/{{item_id}}
HTTP 200

DELETE http://localhost:8080/faq/category/{{category_id}}
HTTP 200

DELETE http://localhost:8080/faq/category/{{category_id}}
HTTP 404
