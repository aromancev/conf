package web

import (
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {
	_, err := graphql.ParseSchema(schema, &Resolver{}, graphql.UseFieldResolvers())
	assert.NoError(t, err)
}
