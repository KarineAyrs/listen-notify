// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package sqlgen

import (
	"context"
)

const listAll = `-- name: ListAll :many
select users.id, users.first_name, users.last_name
from users
`

type ListAllRow struct {
	User User
}

func (q *Queries) ListAll(ctx context.Context) ([]ListAllRow, error) {
	rows, err := q.db.QueryContext(ctx, listAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllRow
	for rows.Next() {
		var i ListAllRow
		if err := rows.Scan(&i.User.ID, &i.User.FirstName, &i.User.LastName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
