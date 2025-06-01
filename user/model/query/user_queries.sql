-- name: GetUserByID :one
SELECT *
FROM Users
WHERE ID = sqlc.arg(id) AND IsDeleted = FALSE;

-- name: GetUserByMobile :one
SELECT *
FROM Users
WHERE Mobile = sqlc.arg(mobile) AND IsDeleted = FALSE;


-- -- name: ListUsers :many
-- SELECT *
-- FROM Users
-- WHERE IsDeleted = FALSE
-- ORDER BY CreateAt DESC;

-- name: UpdateUser :exec
UPDATE Users
SET
    Mobile    = sqlc.arg(mobile),
    Password  = sqlc.arg(password),
    NikeName  = sqlc.arg(nikename),
    Birthday  = sqlc.arg(birthday),
    Gender    = sqlc.arg(gender),
    Role      = sqlc.arg(role),
    UpdateAt  = CURRENT_TIMESTAMP
WHERE
    ID = sqlc.arg(id);


-- name: CreateUser :execresult
INSERT INTO Users (
    Mobile, Password, NikeName, Birthday, Gender, Role
) VALUES (
             sqlc.arg(mobile),
             sqlc.arg(password),
             sqlc.arg(nike_name),
             sqlc.arg(birthday),
             sqlc.arg(gender),
             sqlc.arg(role)
         );

-- name: UpdateUserPassword :exec
UPDATE Users
SET Password = sqlc.arg(password),
    UpdateAt = CURRENT_TIMESTAMP
WHERE ID = sqlc.arg(id) AND IsDeleted = FALSE;

-- name: SoftDeleteUser :exec
UPDATE Users
SET IsDeleted = TRUE,
    DeleteAt = CURRENT_TIMESTAMP
WHERE ID = sqlc.arg(id);

-- name: ListUsers :many
SELECT id, mobile, password, NikeName, birthday, gender, role
FROM Users
WHERE IsDeleted = FALSE
ORDER BY CreateAt DESC
    LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM Users WHERE IsDeleted = FALSE;
