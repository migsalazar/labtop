package modules

import (
	"context"

	"github.com/migsalazar/labtop/internal/model"
)

// Worker owns one collector lifecycle and publishes immutable value updates.
type Worker interface {
	Run(context.Context, chan<- model.ModuleUpdate)
}
