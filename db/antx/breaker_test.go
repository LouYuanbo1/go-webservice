package antx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/skudra-go/ant"
	"github.com/skudra-go/ant/dialect/sqlite"
	_ "modernc.org/sqlite"
)

// user 与 appDB 直接使用原生 ant.Table——antx 只提供熔断中间件，
// 不要求任何包装结构。
type user struct {
	ant.Base
	Name  string
	Email string `ant:"unique"`
	Age   int
}

type appDB struct {
	ant.Context
	Users ant.Table[user]
}

var userFields = ant.Fields[user]()

// stubBreaker 是可手动控制准入结果的 Breaker 实现，用于单测拦截器语义。
type stubBreaker struct {
	allow   error
	success int
	failure int
}

func (s *stubBreaker) Allow() error { return s.allow }

func (s *stubBreaker) MarkSuccess() { s.success++ }

func (s *stubBreaker) MarkFailure() { s.failure++ }

func (s *stubBreaker) Do(ctx context.Context, req func(ctx context.Context) error) error {
	return req(ctx)
}

func (s *stubBreaker) DoWithAcceptable(ctx context.Context, req func(ctx context.Context) error, acceptable func(err error) bool) error {
	return req(ctx)
}

func (s *stubBreaker) DoWithFallback(ctx context.Context, req func(ctx context.Context) error, fallback func(err error) error) error {
	err := req(ctx)
	if err != nil && fallback != nil {
		return fallback(err)
	}
	return err
}

func (s *stubBreaker) GetMetrics() (total, accepts int64, rate float64) { return 0, 0, 0 }

func (s *stubBreaker) Reset() {}

func TestInterceptorAllowsAndRejects(t *testing.T) {
	stub := &stubBreaker{}
	i := &breakerInterceptor{brk: stub, acceptable: DefaultAcceptable}
	ctx := context.Background()

	if _, err := i.Before(ctx, &ant.Statement{SQL: "SELECT 1"}); err != nil {
		t.Fatalf("Before must pass when allow==nil, got %v", err)
	}

	stub.allow = errors.New("breaker open")
	if _, err := i.Before(ctx, &ant.Statement{SQL: "SELECT 1"}); err == nil {
		t.Fatal("Before must reject when breaker denies admission")
	}
}

func TestInterceptorRecordsOutcome(t *testing.T) {
	stub := &stubBreaker{}
	i := &breakerInterceptor{brk: stub, acceptable: DefaultAcceptable}
	ctx := context.Background()

	i.After(ctx, &ant.Statement{}, ant.Outcome{Err: nil})
	i.After(ctx, &ant.Statement{}, ant.Outcome{Err: sql.ErrNoRows})
	i.After(ctx, &ant.Statement{}, ant.Outcome{Err: sql.ErrTxDone})
	i.After(ctx, &ant.Statement{}, ant.Outcome{Err: errors.New("boom")})

	if stub.success != 3 || stub.failure != 1 {
		t.Fatalf("want success=3 failure=1, got success=%d failure=%d", stub.success, stub.failure)
	}
}

func TestWithBreakerEndToEnd(t *testing.T) {
	ctx := context.Background()
	dsn := "file::memory:?cache=shared&_pragma=foreign_keys(1)"

	brk := breaker.NewBreaker()
	db, err := ant.Open[appDB](ctx, sqlite.Dialect, dsn, WithCustomBreaker(brk, nil))
	if err != nil {
		t.Fatal(err)
	}

	const stmt = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL
	)`
	if _, err := ant.Raw[any](db, stmt).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	u := &user{Name: "Bob", Email: "bob@example.com", Age: 25}
	if err := db.Users.Insert(ctx, u); err != nil {
		t.Fatal(err)
	}
	n, err := db.Users.Where(ant.Eq(&userFields.Email, "bob@example.com")).Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("count want 1 with nil err, got n=%d err=%v", n, err)
	}

	total, accepts, rate := brk.GetMetrics()
	if total == 0 || accepts != total || rate != 1 {
		t.Fatalf("want a fully-successful window, got total=%d accepts=%d rate=%v", total, accepts, rate)
	}
}

func TestBreakerSharedAcrossSessions(t *testing.T) {
	ctx := context.Background()
	dsn := "file::memory:?cache=shared&_pragma=foreign_keys(1)"

	brk := breaker.NewBreaker()
	db, err := ant.Open[appDB](ctx, sqlite.Dialect, dsn, WithCustomBreaker(brk, nil))
	if err != nil {
		t.Fatal(err)
	}

	const stmt = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL
	)`
	if _, err := ant.Raw[any](db, stmt).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	if err := db.Users.Insert(ctx, &user{Name: "A", Email: "a@example.com", Age: 1}); err != nil {
		t.Fatal(err)
	}

	before, _, _ := brk.GetMetrics()

	// 每请求一个 Session：同一 Context 派生的 Session 共享同一个熔断实例，
	// 失败统计跨请求持续累积。
	sess := ant.Session(db)
	if err := sess.Users.Insert(ctx, &user{Name: "B", Email: "b@example.com", Age: 2}); err != nil {
		t.Fatal(err)
	}

	after, _, _ := brk.GetMetrics()
	if after <= before {
		t.Fatalf("breaker must be shared across sessions: before=%d after=%d", before, after)
	}
}
