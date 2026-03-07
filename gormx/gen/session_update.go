package gen

import (
	"context"
)

func (g *genSession[T, ID, PT]) Update(ctx context.Context, updateData PT) error {
	return g.Session.Update(ctx, updateData)
}

func (g *genSession[T, ID, PT]) UpdateByStructFilter(ctx context.Context, filter PT, updateData PT) error {
	return g.Session.UpdatesByStructFilter(ctx, filter, updateData)
}

func (g *genSession[T, ID, PT]) UpdateByMapFilter(ctx context.Context, filter map[string]any, updateData map[string]any) error {
	var model T
	ptr := PT(&model)
	return g.Session.UpdatesByMapFilter(ctx, ptr, filter, updateData)
}
