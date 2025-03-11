package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// OTLPClient はOpenTelemetryを使ってログを送信するクライアント
type OTLPClient struct {
	logger         *slog.Logger
	loggerProvider *sdk.LoggerProvider
	attributes     map[string]string
}

// NewOTLPClient は新しいOTLPクライアントを作成します
func NewOTLPClient(ctx context.Context, sid string, endpoint string, attributes map[string]string) (*OTLPClient, error) {
	// リソース情報の設定
	serviceName := "logger"
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Lokiテナント用のヘッダー設定
	headers := map[string]string{
		"X-Scope-OrgID": sid,
	}

	// OTLP HTTP エクスポーターの作成
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithURLPath("/otlp/v1/logs"),
		otlploghttp.WithHeaders(headers),
		otlploghttp.WithInsecure(), // 本番環境ではTLSを使用すること
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	// バッチ処理の設定を最適化
	// 注意: 最新のOpenTelemetry SDKでは設定オプションが異なる可能性があります
	// 実際のSDKのバージョンに合わせて適切なオプションを使用してください
	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithResource(res),
		sdk.WithProcessor(sdk.NewBatchProcessor(exporter)),
	)

	// グローバルなロガープロバイダーとして設定
	global.SetLoggerProvider(loggerProvider)

	// ロガー作成
	logger := otelslog.NewLogger(serviceName)

	return &OTLPClient{
		logger:         logger,
		loggerProvider: loggerProvider,
		attributes:     attributes,
	}, nil
}

// SendLog はログをOpenTelemetryに送信します
func (c *OTLPClient) SendLog(message string) {
	// 空のログは無視
	if message == "" {
		return
	}

	// 属性の設定
	attrs := []attribute.KeyValue{}
	for k, v := range c.attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	// ログレベルに応じた送信 - ここではすべてエラーレベルとして送信
	// TODO: ALB,CloudFrontのログの場合、ログの中からステータスコードを取得して、それにあったログレベルを設定する
	// c.logger.Error(message)
	// c.logger.Info(message)
	c.logger.Warn(message)
	// c.logger.Debug(message)
	// c.logger.Log(message)
}

// FlushLogs はログのバッチを強制的に送信します
func (c *OTLPClient) FlushLogs(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.loggerProvider.ForceFlush(ctx)
}

// Shutdown はOTLPクライアントを正常にシャットダウンします
func (c *OTLPClient) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.loggerProvider.Shutdown(ctx)
}

func main() {
	ctx := context.Background()

	// OpenTelemetryクライアントの設定
	endpoint := "192.168.0.177:31100"

	// カスタム属性の設定
	attributes := map[string]string{
		"environment": "development",
		"application": "log-forwarder",
		"LogType":     "alb",
	}

	// OTLPクライアントの作成
	otlpClient, err := NewOTLPClient(ctx, "fake", endpoint, attributes)
	if err != nil {
		log.Fatalf("Failed to create OTLP client: %v", err)
	}

	// アプリケーション終了時にログを確実に送信し、シャットダウン
	defer func() {
		if err := otlpClient.FlushLogs(5 * time.Second); err != nil {
			log.Printf("Failed to flush logs: %v", err)
		}
		if err := otlpClient.Shutdown(5 * time.Second); err != nil {
			log.Printf("Failed to shutdown client: %v", err)
		}
	}()

	filePath := "/mnt/c/Users/nuts_/Github/Go/LogToOTLPhttp/log.txt"

	for {
		// ファイルの存在確認
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			log.Printf("ファイルは存在しません: %s\n", filePath)
			time.Sleep(300 * time.Second)
			continue
		} else if err != nil {
			log.Printf("エラー: %v\n", err)
			time.Sleep(300 * time.Second)
			continue
		}

		log.Printf("ファイルが存在します: %s - 内容を読み込みます\n", filePath)
		// ファイルを開く
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("ファイルを開けませんでした: %v\n", err)
			time.Sleep(300 * time.Second)
			continue
		}

		// ファイルを行ごとに読み込んで直接OpenTelemetryに送信
		scanner := bufio.NewScanner(file)
		lineCount := 0

		log.Println("ファイルの内容:")
		for scanner.Scan() {
			lineCount++
			line := scanner.Text()

			// ログを直接OpenTelemetryに送信
			otlpClient.SendLog(line)

			log.Printf("行 %d: %s\n", lineCount, line)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("ファイル読み込み中にエラーが発生しました: %v\n", err)
		}

		// 読み込み完了後、バッファに残っているログを確実に送信
		if err := otlpClient.FlushLogs(5 * time.Second); err != nil {
			log.Printf("Failed to flush logs after reading file: %v", err)
		}

		log.Printf("ファイル読み込み完了。合計 %d 行\n", lineCount)

		file.Close()

		// 次の確認まで待機
		time.Sleep(300 * time.Second)
	}
}
