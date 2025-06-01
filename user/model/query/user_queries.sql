-- name: GetUserByID :one
SELECT *
FROM Users
WHERE ID = sqlc.arg(id) AND IsDeleted = FALSE;

-- name: GetUserByMobile :one
SELECT *
FROM Users
WHERE Mobile = sqlc.arg(mobile) AND IsDeleted = FALSE;


-- name: ListUsers :many
SELECT *
FROM Users
WHERE IsDeleted = FALSE
ORDER BY CreateAt DESC;



-- name: CreateUser :exec
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
