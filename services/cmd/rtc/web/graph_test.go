package web

import (
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"
)

func TestResolverImplementsSchema(t *testing.T) {
	_, err := graphql.ParseSchema(gqlSchema, &Resolver{}, graphql.UseFieldResolvers())
	assert.NoError(t, err)
}
