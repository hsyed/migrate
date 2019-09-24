package migrate

import (
	"context"
	"errors"
	"sort"
)

type Changes map[int]string

type Schema struct {
	Name    string
	Changes Changes
}

func validateSchema(schema *Schema) error {
	if schema.Name == "" {
		return errors.New("schema name not set")
	}
	for k, v := range schema.Changes {
		if k <= 0 {
			return errors.New("ordinal must start from 1")
		} else if v == "" {
			return errors.New("no ddl in change")
		}
	}
	return nil
}

type delta struct {
	ordinal int
	ddl     string
}

func filterSortChanges(filter int, changes Changes) []delta {
	ret := make([]delta, 0, len(changes))
	for k, v := range changes {
		if k > filter {
			ret = append(ret, delta{k, v})
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ordinal < ret[j].ordinal })
	return ret
}

type Backend interface {
	Apply(ctx context.Context, schema *Schema) error
	DestroyAndApply(ctx context.Context, schema *Schema) error
}
