package internal

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-webservice/elasticsearchx/config"
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/errors"
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/model"
	"github.com/LouYuanbo1/go-webservice/elasticsearchx/options"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
)

type elasticsearchx[T any, PT model.PointerDocument[T]] struct {
	client            *elasticsearch.TypedClient
	bulkIndexerConfig *config.BulkIndexerConfig
}

func NewElasticsearchX[T any, PT model.PointerDocument[T]](client *elasticsearch.TypedClient, bulkIndexerConfig *config.BulkIndexerConfig) *elasticsearchx[T, PT] {
	return &elasticsearchx[T, PT]{
		client:            client,
		bulkIndexerConfig: bulkIndexerConfig,
	}
}

func (e *elasticsearchx[T, PT]) CreateIndex(ctx context.Context, doc PT) error {
	if doc == nil {
		log.Printf("create index failed : %s", errors.WarnInvalidDocument)
		return nil
	}

	exists, err := e.client.Indices.Exists(doc.Index()).Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrCheckExistence,
			"CreateIndex",
			doc.Index(),
			err,
		)
	}
	if exists {
		log.Printf("create index failed : %s", errors.WarnIndexExist)

		getMappingResponse, err := e.client.Indices.GetMapping().Index(doc.Index()).Do(ctx)
		if err != nil {
			return errors.New(
				errors.ErrGetMapping,
				"CreateIndex",
				doc.Index(),
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
					doc.Index(),
					err,
				)
			} else {
				log.Printf("Index mapping for %s:\n%s", doc.Index(), string(jsonData))
			}
		}
		return nil
	}

	if doc.GetTypeMapping() == nil {
		_, err = e.client.Indices.Create(doc.Index()).Do(ctx)
	} else {
		_, err = e.client.Indices.Create(doc.Index()).Mappings(doc.GetTypeMapping()).Do(ctx)
	}
	if err != nil {
		return errors.New(
			errors.ErrCreateIndex,
			"CreateIndex",
			doc.Index(),
			err,
		)
	}
	return nil
}

func (e *elasticsearchx[T, PT]) GetMapIndexCount(ctx context.Context) (map[string]string, error) {
	resp, err := e.client.Cat.Indices().Do(ctx)
	if err != nil {
		return nil, errors.New(
			errors.ErrGetIndices,
			"GetMapIndexCount",
			"",
			err,
		)
	}
	mapIndiceCount := make(map[string]string, len(resp))
	for _, index := range resp {
		indexName := *index.Index
		// 过滤掉系统索引
		if !strings.HasPrefix(indexName, ".") {
			mapIndiceCount[indexName] = *index.DocsCount
		}
	}
	return mapIndiceCount, nil
}

func (e *elasticsearchx[T, PT]) DeleteIndex(ctx context.Context, index string) error {
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

func (e *elasticsearchx[T, PT]) IndexDoc(ctx context.Context, doc PT) error {
	// 检查文档是否有效
	if doc == nil {
		log.Printf("index document failed : %s", errors.WarnInvalidDocument)
		return nil
	}
	_, err := e.client.Index(doc.Index()).
		Id(doc.GetStringID()).
		Document(doc).
		Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrIndexDocument,
			"IndexDoc",
			doc.Index(),
			err,
		)
	}
	return nil
}

func (e *elasticsearchx[T, PT]) BulkIndexDocs(ctx context.Context, docs []PT, opts ...options.BulkOption) error {
	if len(docs) == 0 {
		log.Printf("bulk index documents failed : %s", errors.WarnEmptyDocumentSlice)
		return nil
	}

	// 构建批量索引器配置
	bulkIndexerConfig := e.bulkIndexerConfigBuilder(opts...)
	bulkIndexerConfig.Client = e.client
	bulkIndexerConfig.Index = docs[0].Index()

	bi, err := esutil.NewBulkIndexer(*bulkIndexerConfig)
	if err != nil {
		return errors.New(
			errors.ErrNewBulkIndexer,
			"BulkIndexDocs",
			docs[0].Index(),
			err,
		)
	}

	// 添加文档到批量索引器
	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			return errors.New(
				errors.ErrMarshalDocument,
				"BulkIndexDocs",
				doc.Index(),
				err,
			)
		}
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",                         // 操作类型：index, create, update, delete
			DocumentID: doc.GetStringID(),               // 文档ID
			Body:       strings.NewReader(string(data)), // 文档内容
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				//fmt.Printf("Successfully indexed document %s\n", item.DocumentID)
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("Error indexing document %s: %s", item.DocumentID, err)
				} else {
					log.Printf("Failed to index document %s: %s", item.DocumentID, res.Error.Reason)
				}
			},
		})
		if err != nil {
			return errors.New(
				errors.ErrBulkIndexDocuments,
				"BulkIndexDocs",
				doc.Index(),
				err,
			)
		}
	}

	// 刷新并关闭批量索引器（确保所有文档都被处理）
	if err := bi.Close(ctx); err != nil {
		log.Printf("Error closing bulk indexer: %s", err)
	}

	// 获取统计信息
	if e.bulkIndexerConfig.Stats {
		stats := bi.Stats()
		log.Printf("Bulk indexing completed:\n")
		log.Printf("Indexed: %d documents\n", stats.NumIndexed)
	}
	return nil
}

func (e *elasticsearchx[T, PT]) GetDoc(ctx context.Context, index string, id string) (PT, error) {
	resp, err := e.client.Get(index, id).Do(ctx)
	if err != nil {
		return nil, errors.New(
			errors.ErrGetDocument,
			"GetDoc",
			index,
			err,
		)
	}
	if !resp.Found {
		log.Printf("get document failed : %s , index: %s, id: %s", errors.WarnDocumentNotFound, index, id)
		return nil, nil
	}
	var doc PT
	err = json.Unmarshal(resp.Source_, doc)
	if err != nil {
		return nil, errors.New(
			errors.ErrUnmarshalDocument,
			"GetDoc",
			index,
			err,
		)
	}
	return doc, nil
}

func (e *elasticsearchx[T, PT]) FindDocsByPages(ctx context.Context, index string, page, size int) ([]PT, error) {
	resp, err := e.client.
		Search().
		Index(index).
		From((page - 1) * size).
		Size(size).
		Do(ctx)
	if err != nil {
		return nil, errors.New(
			errors.ErrFindDocuments,
			"FindDocsByPages",
			index,
			err,
		)
	}
	if resp.Hits.Total.Value == 0 {
		log.Printf("find documents failed : %s , index: %s, page: %d, size: %d", errors.WarnDocumentNotFound, index, page, size)
		return nil, nil
	}
	docs := make([]PT, 0, resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		var doc PT
		err = json.Unmarshal(hit.Source_, doc)
		if err != nil {
			return nil, errors.New(
				errors.ErrUnmarshalDocument,
				"GetDocsByPages",
				index,
				err,
			)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (e *elasticsearchx[T, PT]) CountDocs(ctx context.Context, index string) (int64, error) {
	resp, err := e.client.Count().Index(index).Do(ctx)
	if err != nil {
		return 0, errors.New(
			errors.ErrCountDocument,
			"CountDocs",
			index,
			err,
		)
	}
	return resp.Count, nil
}

func (e *elasticsearchx[T, PT]) UpdateDoc(ctx context.Context, doc PT) error {
	_, err := e.client.Update(doc.Index(), doc.GetStringID()).
		Doc(doc).
		Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrUpdateDocument,
			"UpdateDoc",
			doc.Index(),
			err,
		)
	}
	return nil
}

func (e *elasticsearchx[T, PT]) DeleteDoc(ctx context.Context, index string, id string) error {
	_, err := e.client.Delete(index, id).Do(ctx)
	if err != nil {
		return errors.New(
			errors.ErrDeleteDocument,
			"DeleteDoc",
			index,
			err,
		)
	}
	return nil
}

func (e *elasticsearchx[T, PT]) BulkDeleteDocs(ctx context.Context, index string, ids []string, opts ...options.BulkOption) error {
	if len(ids) == 0 {
		return nil
	}
	// 1. 创建批量索引器配置
	bulk := e.bulkIndexerConfigBuilder(opts...)
	bulk.Client = e.client
	bulk.Index = index

	bi, err := esutil.NewBulkIndexer(*bulk)
	if err != nil {
		return errors.New(
			errors.ErrNewBulkIndexer,
			"BulkDeleteDocs",
			index,
			err,
		)
	}

	// 2. 添加文档到批量索引器
	for _, id := range ids {
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "delete", // 操作类型：index, create, update, delete
			DocumentID: id,       // 文档ID
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				//fmt.Printf("Successfully deleted document %s\n", item.DocumentID)
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("Error deleting document %s: %s", item.DocumentID, err)
				} else {
					log.Printf("Failed to delete document %s: %s", item.DocumentID, res.Error.Reason)
				}
			},
		})
		if err != nil {
			return errors.New(
				errors.ErrBulkDeleteDocuments,
				"BulkDeleteDocs",
				index,
				err,
			)
		}
	}
	// 3. 刷新并关闭批量索引器（确保所有文档都被处理）
	if err := bi.Close(ctx); err != nil {
		log.Printf("Error closing bulk indexer: %s", err)
	}

	// 4. 获取统计信息
	if e.bulkIndexerConfig.Stats {
		stats := bi.Stats()
		log.Printf("Bulk indexing completed:\n")
		log.Printf("Indexed: %d documents\n", stats.NumIndexed)
	}
	return nil
}
