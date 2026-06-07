package models

type Querier func(coordinates Coordinates) (Result, error)
