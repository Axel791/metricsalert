package dto

import "gopkg.in/guregu/null.v4"

type Metrics struct {
	ID    string
	MType string
	Delta null.Int
	Value null.Float
}
