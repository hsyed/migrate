package migrate

import (
	"context"
	"errors"
	"fmt"
)

type Change struct {
	Id string
	Up string
}

type Schema struct {
	Name string
	Changes []Change
}

func validateSchema(schema *Schema) error {
	if schema.Name == "" {
		return errors.New("schema name not set")
	}

	tally := map[string]struct{}{}
	for i, v := range schema.Changes {
		if v.Id == "" {
			return fmt.Errorf("change %d had no id", i+1)
		}
		if _, ok := tally[v.Id]; ok {
			return fmt.Errorf("change id \"%s\" already in use", v.Id)
		} else {
			tally[v.Id] = struct{}{}
		}
	}
	return nil
}

type Backend interface {
	Apply(ctx context.Context, schema *Schema) error
	DestroyAndApply(ctx context.Context, schema *Schema) error
}
