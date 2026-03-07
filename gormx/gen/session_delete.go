package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) DeleteByID(ctx context.Context, id ID) error {
	if IsZero(id) {
		log.Printf("delete by id failed : %s", gormx.WarnInvalidID)
		return nil
	}

	var model T
	ptr := PT(&model)
	return g.Session.DeleteByID(ctx, ptr, id)
}

func (g *genSession[T, ID, PT]) DeleteByIDs(ctx context.Context, ids []ID) error {
	if len(ids) == 0 {
		log.Printf("delete by ids failed : %s", gormx.WarnEmptyIDSlice)
		return nil
	}
	for _, id := range ids {
		if IsZero(id) {
			log.Printf("delete by ids failed, index: %v : %s", id, gormx.WarnInvalidID)
			return nil
		}
	}

	var model T
	ptr := PT(&model)
	return g.Session.DeleteByIDs(ctx, ptr, ids)
}

func (g *genSession[T, ID, PT]) DeleteByStructFilter(ctx context.Context, filter PT) error {
	var model T
	ptr := PT(&model)
	return g.Session.DeleteByStructFilter(ctx, ptr, filter)
}

func (g *genSession[T, ID, PT]) DeleteByMapFilter(ctx context.Context, filter map[string]any) error {
	var model T
	ptr := PT(&model)
	return g.Session.DeleteByMapFilter(ctx, ptr, filter)
}
