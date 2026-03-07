package gen

import (
	"context"
	"log"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (g *genSession[T, ID, PT]) GetByID(ctx context.Context, id ID) (PT, error) {
	if IsZero(id) {
		log.Printf("get by id failed : %s", gormx.WarnInvalidID)
		return nil, nil
	}

	var model T
	ptr := PT(&model)
	if err := g.Session.GetByID(ctx, ptr, id); err != nil {
		return nil, err
	}
	return ptr, nil
}

func (g *genSession[T, ID, PT]) GetByStructFilter(ctx context.Context, filter PT) (PT, error) {
	var model T
	ptr := PT(&model)
	if err := g.Session.GetByStructFilter(ctx, ptr, filter); err != nil {
		return nil, err
	}
	return ptr, nil
}

func (g *genSession[T, ID, PT]) GetByMapFilter(ctx context.Context, filter map[string]any) (PT, error) {
	var model T
	ptr := PT(&model)
	if err := g.Session.GetByMapFilter(ctx, ptr, filter); err != nil {
		return nil, err
	}
	return ptr, nil
}
