package graph

import (
	"github.com/graphql-go/graphql"
	"github.com/toluhikay/go-react/internal/models"
)

type Graph struct {
	Movies      []*models.Movie
	QueryString string
	Config      graphql.SchemaConfig
	// fields      graphql.Fields
	// movieType   *graphql.Object
}

// create a new constructor to return the Graph

// s
