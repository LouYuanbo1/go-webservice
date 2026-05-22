package elasticsearchx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-webservice/errorx"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
)

type Elasticsearchx struct {
	client *elasticsearch.TypedClient
}

func NewElasticsearchX(client *elasticsearch.TypedClient) *Elasticsearchx {
	return &Elasticsearchx{
		client: client,
	}
}

func (e *Elasticsearchx) CreateIndex[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error {
	index := doc.Index()

	exist, err := e.client.Indices.Exists(index).Do(ctx)
	if err != nil {
		return errorx.New(
			ErrCheckExistence,
			"elasticsearchx",
			fmt.Sprintf("CreateIndex[%s]", index),
			err,
		)
	}
	if exist {
		log.Printf("create index failed : %s", WarnIndexExist)

		getMappingResponse, err := e.client.Indices.GetMapping().Index(index).Do(ctx)
		if err != nil {
			return errorx.New(
				ErrGetMapping,
				"elasticsearchx",
				fmt.Sprintf("CreateIndex[%s]", index),
				err,
			)
		}
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
			return errorx.New(
				ErrMarshalMapping,
				"elasticsearchx",
				fmt.Sprintf("CreateIndex[%s]", index),
				err,
			)
		}

		log.Printf("Index mapping for %s:\n%s", index, string(jsonData))
		return nil
	}

	if doc.GetTypeMapping() == nil {
		_, err = e.client.Indices.Create(index).Do(ctx)
	} else {
		_, err = e.client.Indices.Create(index).Mappings(doc.GetTypeMapping()).Do(ctx)
	}
	if err != nil {
		return errorx.New(
			ErrCreateIndex,
			"elasticsearchx",
			fmt.Sprintf("CreateIndex[%s]", index),
			err,
		)
	}
	return nil
}

func (e *Elasticsearchx) GetIndices(ctx context.Context) (mapIndexCount map[string]string, err error) {
	resp, err := e.client.Cat.Indices().Do(ctx)
	if err != nil {
		return nil, errorx.New(
			ErrGetIndices,
			"elasticsearchx",
			"GetIndices",
			err,
		)
	}
	mapIndexCount = make(map[string]string, len(resp))
	for _, item := range resp {
		mapIndexCount[*item.Index] = *item.DocsCount
	}
	return mapIndexCount, nil
}

func (e *Elasticsearchx) DeleteIndex[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error {
	index := doc.Index()

	_, err := e.client.Indices.Delete(index).Do(ctx)
	if err != nil {
		return errorx.New(
			ErrDeleteIndex,
			"elasticsearchx",
			fmt.Sprintf("DeleteIndex[%s]", index),
			err,
		)
	}
	return nil
}

func (e *Elasticsearchx) IndexDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error {
	// 检查文档是否有效
	if doc == nil {
		log.Printf("index document failed : %s", WarnInvalidDocument)
		return nil
	}
	index := doc.Index()

	_, err := e.client.Index(index).
		Id(doc.GetStringID()).
		Document(doc).
		Do(ctx)
	if err != nil {
		return errorx.New(
			ErrIndexDocument,
			"elasticsearchx",
			fmt.Sprintf("IndexDoc[%s]", index),
			err,
		)
	}
	return nil
}

func (e *Elasticsearchx) BulkIndexDocs[T any, PT PointerDocument[T]](ctx context.Context, docs []PT, cfg esutil.BulkIndexerConfig, stats bool) error {
	if len(docs) == 0 {
		log.Printf("bulk index documents failed : %s", WarnEmptyDocumentSlice)
		return nil
	}
	doc := PT(new(T))
	index := doc.Index()

	bi, err := esutil.NewBulkIndexer(cfg)
	if err != nil {
		return errorx.New(
			ErrNewBulkIndexer,
			"elasticsearchx",
			fmt.Sprintf("BulkIndexDocs[%s]", index),
			err,
		)
	}

	// 添加文档到批量索引器
	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			return errorx.New(
				ErrMarshalDocument,
				"elasticsearchx",
				fmt.Sprintf("BulkIndexDocs[%s]", index),
				err,
			)
		}
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",                         // 操作类型：index, create, update, delete
			DocumentID: doc.GetStringID(),               // 文档ID
			Index:      index,                           // 索引名称
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
			return errorx.New(
				ErrBulkIndexDocuments,
				"elasticsearchx",
				fmt.Sprintf("BulkIndexDocs[%s]", index),
				err,
			)
		}
	}

	// 刷新并关闭批量索引器（确保所有文档都被处理）
	if err := bi.Close(ctx); err != nil {
		log.Printf("Error closing bulk indexer: %s", err)
	}

	// 获取统计信息
	if stats {
		stats := bi.Stats()
		log.Printf("Bulk indexing completed:\n")
		log.Printf("Indexed: %d documents\n", stats.NumIndexed)
	}
	return nil
}

func (e *Elasticsearchx) GetDoc[T any, PT PointerDocument[T]](ctx context.Context, id string) (PT, error) {
	doc := PT(new(T))
	index := doc.Index()

	resp, err := e.client.Get(index, id).Do(ctx)
	if err != nil {
		return nil, errorx.New(
			ErrGetDocument,
			"elasticsearchx",
			fmt.Sprintf("GetDoc[%s]", index),
			err,
		)
	}
	if !resp.Found {
		log.Printf("get document failed : %s , index: %s, id: %s", WarnDocumentNotFound, index, id)
		return nil, nil
	}
	err = json.Unmarshal(resp.Source_, doc)
	if err != nil {
		return nil, errorx.New(
			ErrUnmarshalDocument,
			"elasticsearchx",
			fmt.Sprintf("GetDoc[%s]", index),
			err,
		)
	}
	return doc, nil
}

func (e *Elasticsearchx) FindDocsByPages[T any, PT PointerDocument[T]](ctx context.Context, page, size int) (*[]T, error) {
	doc := PT(new(T))
	index := doc.Index()

	resp, err := e.client.
		Search().
		Index(index).
		From((page - 1) * size).
		Size(size).
		Do(ctx)
	if err != nil {
		return nil, errorx.New(
			ErrFindDocuments,
			"elasticsearchx",
			fmt.Sprintf("FindDocsByPages[%s]", index),
			err,
		)
	}
	if resp.Hits.Total.Value == 0 {
		log.Printf("find documents failed : %s , index: %s, page: %d, size: %d", WarnDocumentNotFound, index, page, size)
		return nil, nil
	}
	docs := make([]T, 0, resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		var newDoc T
		err = json.Unmarshal(hit.Source_, &newDoc)
		if err != nil {
			return nil, errorx.New(
				ErrUnmarshalDocument,
				"elasticsearchx",
				fmt.Sprintf("FindDocsByPages[%s]", index),
				err,
			)
		}
		docs = append(docs, newDoc)
	}
	return &docs, nil
}

func (e *Elasticsearchx) CountDocs[T any, PT PointerDocument[T]](ctx context.Context) (int64, error) {
	doc := PT(new(T))
	index := doc.Index()

	resp, err := e.client.Count().Index(index).Do(ctx)
	if err != nil {
		return 0, errorx.New(
			ErrCountDocument,
			"elasticsearchx",
			fmt.Sprintf("CountDocs[%s]", index),
			err,
		)
	}
	return resp.Count, nil
}

func (e *Elasticsearchx) UpdateDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error {
	// 检查文档是否有效
	if doc == nil {
		log.Printf("update document failed : %s", WarnInvalidDocument)
		return nil
	}
	index := doc.Index()

	_, err := e.client.Update(index, doc.GetStringID()).
		Doc(doc).
		Do(ctx)
	if err != nil {
		return errorx.New(
			ErrUpdateDocument,
			"elasticsearchx",
			fmt.Sprintf("UpdateDoc[%s]", index),
			err,
		)
	}
	return nil
}

func (e *Elasticsearchx) DeleteDoc[T any, PT PointerDocument[T]](ctx context.Context, doc PT) error {
	index := doc.Index()

	_, err := e.client.Delete(index, doc.GetStringID()).Do(ctx)
	if err != nil {
		return errorx.New(
			ErrDeleteDocument,
			"elasticsearchx",
			fmt.Sprintf("DeleteDoc[%s]", index),
			err,
		)
	}
	return nil
}

func (e *Elasticsearchx) BulkDeleteDocs[T any, PT PointerDocument[T]](ctx context.Context, ids []string, cfg esutil.BulkIndexerConfig, stats bool) error {
	if len(ids) == 0 {
		log.Printf("bulk delete documents failed : %s", WarnEmptyDocumentSlice)
		return nil
	}

	doc := PT(new(T))
	index := doc.Index()

	bi, err := esutil.NewBulkIndexer(cfg)
	if err != nil {
		return errorx.New(
			ErrNewBulkIndexer,
			"elasticsearchx",
			fmt.Sprintf("BulkDeleteDocs[%s]", index),
			err,
		)
	}

	// 2. 添加文档到批量索引器
	for _, id := range ids {
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "delete", // 操作类型：index, create, update, delete
			DocumentID: id,       // 文档ID
			Index:      index,    // 索引名称
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
			return errorx.New(
				ErrBulkDeleteDocuments,
				"elasticsearchx",
				fmt.Sprintf("BulkDeleteDocs[%s]", index),
				err,
			)
		}
	}
	// 3. 刷新并关闭批量索引器（确保所有文档都被处理）
	if err := bi.Close(ctx); err != nil {
		log.Printf("Error closing bulk indexer: %s", err)
	}

	// 4. 获取统计信息
	if stats {
		stats := bi.Stats()
		log.Printf("Bulk indexing completed:\n")
		log.Printf("Indexed: %d documents\n", stats.NumIndexed)
	}
	return nil
}
