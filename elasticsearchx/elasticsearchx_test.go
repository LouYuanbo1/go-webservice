package elasticsearchx

/*
import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/stretchr/testify/assert"
	elasticsearchcontainer "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

// TestDocument 用于测试的文档类型
type TestDocument struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (d *TestDocument) Index() string {
	return "test_index"
}

func (d *TestDocument) GetStringID() string {
	return d.ID
}

func (d *TestDocument) GetTypeMapping() *types.TypeMapping {
	return nil
}

func setupElasticsearch(ctx context.Context, t *testing.T) (*Elasticsearchx, *elasticsearch.TypedClient, func()) {
	// 启动 Elasticsearch 容器（使用 7.x 版本以避免安全认证问题）
	container, err := elasticsearchcontainer.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:7.17.0")
	if err != nil {
		t.Fatalf("Failed to start Elasticsearch container: %v", err)
	}

	// 获取容器端点（使用 HTTP）
	endpoint, err := container.Endpoint(ctx, "http")
	if err != nil {
		t.Fatalf("Failed to get container endpoint: %v", err)
	}

	// 创建 Elasticsearch 客户端
	cfg := elasticsearch.Config{
		Addresses: []string{endpoint},
	}
	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// 创建 Elasticsearchx 实例
	esx := NewElasticsearchX(client)

	// 清理函数
	cleanup := func() {
		client.Close(context.Background())
		container.Terminate(context.Background())
	}

	return esx, client, cleanup
}

func TestElasticsearchx_CreateIndex(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 验证索引存在
	indices, err := esx.GetIndices(ctx)
	assert.NoError(t, err)
	assert.Contains(t, indices, "test_index")
}

func TestElasticsearchx_CreateIndex_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}

	// 第一次创建索引
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 第二次创建相同索引（应该不报错，只是警告）
	err = esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)
}

func TestElasticsearchx_IndexDoc(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 先创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 索引文档
	err = esx.IndexDoc(ctx, doc)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 验证文档已索引
	retrieved, err := esx.GetDoc[TestDocument](ctx, "1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test", retrieved.Name)
	assert.Equal(t, 30, retrieved.Age)
}

func TestElasticsearchx_IndexDoc_NilDocument(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 索引 nil 文档（应该不报错）
	err := esx.IndexDoc[TestDocument](ctx, nil)
	assert.NoError(t, err)
}

func TestElasticsearchx_GetDoc(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引并索引文档
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)
	err = esx.IndexDoc(ctx, doc)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 获取文档
	retrieved, err := esx.GetDoc[TestDocument](ctx, "1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test", retrieved.Name)
	assert.Equal(t, 30, retrieved.Age)
}

func TestElasticsearchx_GetDoc_NotFound(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 获取不存在的文档
	retrieved, err := esx.GetDoc[TestDocument](ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestElasticsearchx_UpdateDoc(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引并索引文档
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)
	err = esx.IndexDoc(ctx, doc)
	assert.NoError(t, err)

	// 更新文档
	updatedDoc := &TestDocument{ID: "1", Name: "Updated", Age: 35}
	err = esx.UpdateDoc(ctx, updatedDoc)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 验证文档已更新
	retrieved, err := esx.GetDoc[TestDocument](ctx, "1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Updated", retrieved.Name)
	assert.Equal(t, 35, retrieved.Age)
}

func TestElasticsearchx_DeleteDoc(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引并索引文档
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)
	err = esx.IndexDoc(ctx, doc)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 删除文档
	err = esx.DeleteDoc(ctx, doc)
	assert.NoError(t, err)

	// 验证文档已删除
	retrieved, err := esx.GetDoc[TestDocument](ctx, "1")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestElasticsearchx_CountDocs(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 索引多个文档
	for i := 1; i <= 5; i++ {
		err := esx.IndexDoc(ctx, &TestDocument{ID: string(rune('0' + i)), Name: "Test", Age: 30})
		assert.NoError(t, err)
	}

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 统计文档数量
	count, err := esx.CountDocs[TestDocument](ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestElasticsearchx_FindDocsByPages(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 索引多个文档
	for i := 1; i <= 10; i++ {
		err := esx.IndexDoc(ctx, &TestDocument{ID: string(rune('0' + i)), Name: "Test", Age: 30})
		assert.NoError(t, err)
	}

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 分页查询第一页
	docs, err := esx.FindDocsByPages[TestDocument](ctx, 1, 3)
	assert.NoError(t, err)
	assert.NotNil(t, docs)
	assert.Len(t, *docs, 3)
}

func TestElasticsearchx_BulkIndexDocs(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 准备批量文档
	docs := []*TestDocument{
		{ID: "1", Name: "Test1", Age: 20},
		{ID: "2", Name: "Test2", Age: 25},
		{ID: "3", Name: "Test3", Age: 30},
	}

	// 批量索引文档
	cfg := esutil.BulkIndexerConfig{
		Client: client,
	}
	err = esx.BulkIndexDocs(ctx, docs, cfg, false)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 验证文档数量
	count, err := esx.CountDocs[TestDocument](ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestElasticsearchx_BulkIndexDocs_EmptySlice(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 批量索引空切片（应该不报错）
	cfg := esutil.BulkIndexerConfig{}
	err := esx.BulkIndexDocs[TestDocument](ctx, []*TestDocument{}, cfg, false)
	assert.NoError(t, err)
}

func TestElasticsearchx_BulkDeleteDocs(t *testing.T) {
	ctx := context.Background()
	esx, client, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引并索引文档
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	for i := 1; i <= 5; i++ {
		err := esx.IndexDoc(ctx, &TestDocument{ID: string(rune('0' + i)), Name: "Test", Age: 30})
		assert.NoError(t, err)
	}

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 批量删除文档
	cfg := esutil.BulkIndexerConfig{
		Client: client,
	}
	err = esx.BulkDeleteDocs[TestDocument](ctx, []string{"1", "2", "3"}, cfg, false)
	assert.NoError(t, err)

	// 刷新索引
	_, err = client.Indices.Refresh().Index("test_index").Do(ctx)
	assert.NoError(t, err)

	// 验证剩余文档数量
	count, err := esx.CountDocs[TestDocument](ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestElasticsearchx_BulkDeleteDocs_EmptySlice(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 批量删除空切片（应该不报错）
	cfg := esutil.BulkIndexerConfig{}
	err := esx.BulkDeleteDocs[TestDocument](ctx, []string{}, cfg, false)
	assert.NoError(t, err)
}

func TestElasticsearchx_DeleteIndex(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 创建索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err := esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 删除索引
	err = esx.DeleteIndex(ctx, doc)
	assert.NoError(t, err)

	// 验证索引不存在
	indices, err := esx.GetIndices(ctx)
	assert.NoError(t, err)
	assert.NotContains(t, indices, "test_index")
}

func TestElasticsearchx_GetIndices(t *testing.T) {
	ctx := context.Background()
	esx, _, cleanup := setupElasticsearch(ctx, t)
	defer cleanup()

	// 获取初始索引列表
	indices, err := esx.GetIndices(ctx)
	assert.NoError(t, err)

	// 创建测试索引
	doc := &TestDocument{ID: "1", Name: "Test", Age: 30}
	err = esx.CreateIndex(ctx, doc)
	assert.NoError(t, err)

	// 再次获取索引列表，验证新索引已创建
	indices, err = esx.GetIndices(ctx)
	assert.NoError(t, err)
	assert.Contains(t, indices, "test_index")
}
*/
