package notes

import "github.com/sudachen/go-ml/tables"

type Chart interface {
	Plot(*tables.Table)
}
