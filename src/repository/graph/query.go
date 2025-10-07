package graph

// QueryBuilder defines a function that returns a query string and bind variables.
type QueryBuilder func() (string, map[string]interface{})
