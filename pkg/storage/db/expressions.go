package db

import (
	"errors"
	"reflect"
	"strings"

	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/goqu/v9/exp"
)

var ErrTagNotFound error = errors.New("tag not found")

func parseFilter(input string) ([]exp.Expression, error) {
	params, err := filter.Parse(input)
	if err != nil {
		return nil, err
	}

	expressions := []exp.Expression{}
	for _, param := range params {
		operator, err := parseOperator(param.Operator)
		if err != nil {
			return nil, err
		}

		field, err := findField(param.Attribute)
		if err != nil {
			return nil, err
		}

		expressions = append(expressions, exp.Ex{
			usersTable + "." + field: exp.Op{operator: param.Value},
		})
	}

	return expressions, nil
}

func parseOperator(operator filter.Operator) (string, error) {
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

func findField(input string) (string, error) {
	entityValue := reflect.TypeOf(entity.User{})
	datasetValue := reflect.TypeOf(sqlUser{})

	for i := 0; i < entityValue.NumField(); i++ {
		entityField := entityValue.Field(i)
		tag, _, _ := strings.Cut(entityField.Tag.Get("json"), ",")

		if tag != input {
			continue
		}

		datasetField, ok := datasetValue.FieldByName(entityField.Name)
		if !ok {
			return "", ErrTagNotFound
		}

		datasetTag, ok := datasetField.Tag.Lookup("db")
		if !ok {
			return "", ErrTagNotFound
		}

		return datasetTag, nil
	}

	return "", ErrTagNotFound
}
