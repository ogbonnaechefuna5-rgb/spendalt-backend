package core

import "database/sql"

// ScanRows iterates rows, calling scan to get the destination pointers for each item.
func ScanRows[T any](rows *sql.Rows, scan func(*T) []interface{}) ([]*T, error) {
	defer rows.Close()
	var results []*T
	for rows.Next() {
		item := new(T)
		if err := rows.Scan(scan(item)...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}
