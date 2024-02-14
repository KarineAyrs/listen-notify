-- name: ListAll :many
select sqlc.embed(users)
from users;