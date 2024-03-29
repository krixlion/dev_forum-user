package cockroach

import (
	"errors"
	"reflect"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/krixlion/dev_forum-lib/filter"
)

var ErrTagNotFound error = errors.New("tag not found")

// filterToSqlExp converts filter.Filter into goqu.Expression to use with goqu SQL builder.
func filterToSqlExp(params filter.Filter) ([]exp.Expression, error) {
	expressions := make([]exp.Expression, 0, len(params))
	for _, param := range params {
		operator, err := matchOperator(param.Operator)
		if err != nil {
			return nil, err
		}

		if err := verifyField(param.Attribute); err != nil {
			return nil, err
		}

		expressions = append(expressions, exp.Ex{
			usersTable + "." + param.Attribute: exp.Op{operator: param.Value},
		})
	}

	return expressions, nil
}

// matchOperator returns a goqu operator corresponding to the filter.Operator specs.
func matchOperator(operator filter.Operator) (string, error) {
	switch operator {
	case filter.Equal:
		return "eq", nil
	case filter.NotEqual:
		return "neq", nil
	case filter.GreaterThan:
		return "gt", nil
	case filter.GreaterThanOrEqual:
		return "gte", nil
	case filter.LesserThan:
		return "lt", nil
	case filter.LesserThanOrEqual:
		return "lte", nil
	default:
		return "", errors.New("invalid operator")
	}
}

// verifyField check whether provided field name is one of dataset's fields
// based on associated tags.
func verifyField(input string) error {
	datasetType := reflect.TypeOf(userDataset{})

	for i := 0; i < datasetType.NumField(); i++ {
		field := datasetType.Field(i)
		tag, ok := field.Tag.Lookup("db")
		if !ok {
			return ErrTagNotFound
		}

		if tag == input {
			return nil
		}
	}

	return ErrTagNotFound
}
