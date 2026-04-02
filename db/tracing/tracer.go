package tracing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer 数据库操作追踪器
type Tracer struct {
	tracer trace.Tracer
	config *TracerConfig
}

// TracerConfig 追踪器配置
type TracerConfig struct {
	ServiceName string // 服务名称
	DBName      string // 数据库名称
	DBSystem    string // 数据库类型（mysql/postgresql/sqlite）
}

// NewTracer 创建数据库追踪器
func NewTracer(config *TracerConfig) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(config.ServiceName),
		config: config,
	}
}

// TraceQuery 追踪查询操作
func (t *Tracer) TraceQuery(ctx context.Context, query string, args []interface{}) (context.Context, error) {
	// 创建 Span
	spanName := fmt.Sprintf("DB Query: %s", t.getOperationName(query))
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// 设置属性
	span.SetAttributes(
		attribute.String("db.system", t.config.DBSystem),
		attribute.String("db.name", t.config.DBName),
		attribute.String("db.statement", query),
		attribute.StringSlice("db.args", t.argsToString(args)),
	)

	// 返回包含 Span 的 context
	return ctx, nil
}

// RecordError 记录错误
func (t *Tracer) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
	}
}

// RecordAffectedRows 记录影响的行数
func (t *Tracer) RecordAffectedRows(ctx context.Context, rows int64) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.Int64("db.rows_affected", rows))
	}
}

// getOperationName 从 SQL 获取操作名称
func (t *Tracer) getOperationName(sql string) string {
	if len(sql) < 6 {
		return "UNKNOWN"
	}

	prefix := sql[:6]
	switch prefix {
	case "SELECT":
		return "SELECT"
	case "INSERT":
		return "INSERT"
	case "UPDATE":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "OTHER"
	}
}

// argsToString 将参数转换为字符串切片
func (t *Tracer) argsToString(args []interface{}) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = fmt.Sprintf("%v", arg)
	}
	return result
}

// WithTrace 在查询中启用追踪
func WithTrace(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	// 获取追踪器（从全局或上下文中）
	tracer := getTracerFromContext(ctx)

	if tracer != nil {
		var err error
		ctx, err = tracer.TraceQuery(ctx, query, args)
		if err != nil {
			return nil, err
		}

		defer func(start time.Time) {
			duration := time.Since(start)
			span := trace.SpanFromContext(ctx)
			if span.IsRecording() {
				span.SetAttributes(attribute.Int64("db.duration_ms", duration.Milliseconds()))
			}
		}(time.Now())
	}

	// 执行实际查询
	return db.QueryContext(ctx, query, args...)
}

// ExecWithTrace 在执行中启用追踪
func ExecWithTrace(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	tracer := getTracerFromContext(ctx)

	if tracer != nil {
		var err error
		ctx, err = tracer.TraceQuery(ctx, query, args)
		if err != nil {
			return nil, err
		}

		defer func(start time.Time) {
			duration := time.Since(start)
			span := trace.SpanFromContext(ctx)
			if span.IsRecording() {
				span.SetAttributes(attribute.Int64("db.duration_ms", duration.Milliseconds()))
			}
		}(time.Now())
	}

	// 执行实际操作
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		if tracer != nil {
			tracer.RecordError(ctx, err)
		}
		return nil, err
	}

	// 记录影响的行数
	if tracer != nil {
		rows, _ := result.RowsAffected()
		tracer.RecordAffectedRows(ctx, rows)
	}

	return result, nil
}

// contextKey 上下文键类型
type contextKey string

const tracerKey contextKey = "db_tracer"

// ContextWithTracer 将追踪器存入上下文
func ContextWithTracer(ctx context.Context, tracer *Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

// getTracerFromContext 从上下文获取追踪器
func getTracerFromContext(ctx context.Context) *Tracer {
	if tracer, ok := ctx.Value(tracerKey).(*Tracer); ok {
		return tracer
	}
	return nil
}
