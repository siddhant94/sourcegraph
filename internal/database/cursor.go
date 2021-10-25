package database

// A Cursor for efficient index based pagination through large result sets.
type Cursor struct {
	// Column contains the relevant column for cursor-based pagination (e.g. "name")
	Column string
	// Value contains the relevant value for cursor-based pagination (e.g. "Zaphod").
	Value string
	// Direction contains the comparison for cursor-based pagination, all possible values are: next, prev.
	Direction string
}
