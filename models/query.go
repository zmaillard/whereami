package models

import "context"

type Querier func(ctx context.Context, coordinates Coordinates) (Result, error)
