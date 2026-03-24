package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) GetByID(ctx context.Context, dest PT, id ID) error {
	if IsZero(id) {
		log.Printf("get by id failed : %s", gormx.WarnInvalidID)
		return nil
	}
	return g.Session.GetByID(ctx, dest, id)
}

func (g *genSession[T, ID, PT]) GetByStructFilter(ctx context.Context, dest PT, filter PT) error {
	return g.Session.GetByStructFilter(ctx, dest, filter)
}

func (g *genSession[T, ID, PT]) GetByMapFilter(ctx context.Context, dest PT, filter map[string]any) error {
	return g.Session.GetByMapFilter(ctx, dest, filter)
}
