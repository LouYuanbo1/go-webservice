package internal

import (
	"context"
	"encoding/json"
	"log"

	"github.com/LouYuanbo1/go-webservice/elasticsearchx/errors"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

type elasticsearchx struct {
	client *elasticsearch.TypedClient
}

func NewElasticsearchX(client *elasticsearch.TypedClient) *elasticsearchx {
	return &elasticsearchx{client: client}
}

func (e *elasticsearchx) CreateIndex(ctx context.Context, index string, mapping ...*types.TypeMapping) error {
	exists, err := e.client.Indices.Exists(index).Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrCheckExistence,
			"CreateIndex",
			index,
			err,
		)
	}
	if exists {
		log.Printf("Index %s already exists, skip create", index)
		getMappingResponse, err := e.client.Indices.GetMapping().Index(index).Do(ctx)
		if err != nil {
			return errors.New(
				errors.ErrGetMapping,
				"CreateIndex",
				index,
				err,
			)
		} else {
			// 将mapping转换为JSON格式打印
			//json.MarshalIndent
			// 格式化格式：生成人类可读的、带缩进和换行的 JSON
			// 适合场景：日志记录、调试、配置文件、人类阅读等
			// 第一个参数 "" (prefix) - 行前缀
			// 作用：指定每一行 JSON 数据开头的前缀字符串
			// 第二个参数 " " (indent) - 缩进字符
			// 作用：指定每一级嵌套使用的缩进字符串
			jsonData, err := json.MarshalIndent(getMappingResponse, "", "  ")
			if err != nil {
				return errors.New(
					errors.ErrMarshalMapping,
					"CreateIndex",
					index,
					err,
				)
			} else {
				log.Printf("Index mapping for %s:\n%s", index, string(jsonData))
			}
		}
		return nil
	}

	if mapping == nil {
		_, err = e.client.Indices.Create(index).Do(ctx)
	} else {
		_, err = e.client.Indices.Create(index).Mappings(mapping[0]).Do(ctx)
	}
	if err != nil {
		return errors.New(
			errors.ErrCreateIndex,
			"CreateIndex",
			index,
			err,
		)
	}
	return nil
}

func (e *elasticsearchx) DeleteIndex(ctx context.Context, index string) error {
	_, err := e.client.Indices.Delete(index).Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrDeleteIndex,
			"DeleteIndex",
			index,
			err,
		)
	}
	return nil
}
