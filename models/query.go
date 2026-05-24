package models

import "net/http"

type Querier func(client *http.Client, coordinates Coordinates) (Result, error)
