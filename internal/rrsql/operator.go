package rrsql

import "errors"

var allowOperator = map[string]string{
	"$gte": ">=",
	"$gt":  ">",
	"$lte": "<=",
	"$lt":  "<",
	"$neq": "!=",
	"$eq":  "=",
	"$in":  "IN",
	"$nin": "NOT IN",
}

// OperatorCoverter converts between query operators and MySQL operators
func OperatorCoverter(op string) (r string, err error) {
	if result, ok := allowOperator[op]; ok {
		r = result
	} else {
		return "", errors.New("invalid operator")
	}
	return r, nil
}
