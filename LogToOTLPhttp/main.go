package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	QueueURL     string = getEnvironmentVariable("SQS_QUEUE_URL", "https://sqs.ap-northeast-1.amazonaws.com/637423497892/lee-test-standard-queue-for-alb-cloudfront-logs")
	otlpEndpoint string = getEnvironmentVariable("OTLP_ENDPOINT", "10.1.2.218:31100")
)

func getEnvironmentVariable(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type S3EventNotification struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
				Arn  string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key  string `json:"key"`
				Size int    `json:"size"`
				ETag string `json:"eTag"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// OTLPClient はOpenTelemetryを使ってログを送信するクライアント
type OTLPClient struct {
	logger         *slog.Logger
	loggerProvider *sdk.LoggerProvider
	sid            string
	attributes     map[string]string
}

func NewOTLPClient(ctx context.Context, sid string, attributes map[string]string) (*OTLPClient, error) {
	// リソース情報の設定
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(sid),
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
		otlploghttp.WithEndpoint(otlpEndpoint),
		otlploghttp.WithURLPath("/otlp/v1/logs"),
		otlploghttp.WithHeaders(headers),
		otlploghttp.WithInsecure(), // 本番環境ではTLSを使用すること
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithResource(res),
		sdk.WithProcessor(sdk.NewBatchProcessor(exporter)),
	)

	// グローバルなロガープロバイダーとして設定
	global.SetLoggerProvider(loggerProvider)

	// ロガー作成
	logger := otelslog.NewLogger("otlphttp-logger") // "scope_name"というStructured metadataとして設定される

	return &OTLPClient{
		logger:         logger,
		loggerProvider: loggerProvider,
		sid:            sid,
		attributes:     attributes,
	}, nil
}

// SendLog はログをOpenTelemetryに送信します
func (c *OTLPClient) SendLog(cfg aws.Config, bucketName string, object string) {
	// ログを確実に送信し、シャットダウン
	defer func() {
		if err := c.FlushLogs(10 * time.Second); err != nil {
			log.Printf("Failed to flush logs: %v", err)
		}
		if err := c.Shutdown(10 * time.Second); err != nil {
			log.Printf("Failed to shutdown client: %v", err)
		}
	}()

	fmt.Printf("SendLog Goroutine started (%s)\n", object)
	startTime := time.Now()

	s3_client := s3.NewFromConfig(cfg)
	result, err := s3_client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(object),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", object, bucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, object, err)
		}
		return
	}
	defer result.Body.Close()
	splitFileName := strings.Split(object, "/")
	gzipFileName := splitFileName[len(splitFileName)-1]

	file, err := os.Create(gzipFileName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", gzipFileName, err)
		return
	}

	// S3から取得したデータをファイルに書き込む
	_, err = io.Copy(file, result.Body)
	if err != nil {
		log.Printf("Couldn't copy S3 object to file. Here's why: %v\n", err)
		file.Close()
		return
	}
	file.Close()

	gzipFile, err := os.Open(gzipFileName)
	if err != nil {
		log.Printf("Could not open file %s: %v\n", gzipFileName, err)
		return
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		log.Printf("Could not create gzip reader: %v\n", err)
		return
	}
	defer gzipReader.Close()

	// 出力先ファイルを作成
	outputFile, err := os.Create(strings.TrimSuffix(gzipFileName, ".gz"))
	if err != nil {
		log.Printf("Could not create output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// 解凍したデータを出力ファイルにコピー
	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		log.Printf("Could not copy data to output file: %v\n", err)
		return
	}

	// ファイルポインタを先頭に戻す
	_, err = outputFile.Seek(0, 0)
	if err != nil {
		log.Printf("Could not reset file pointer: %v\n", err)
		return
	}

	// 処理完了後、一時ファイルを削除
	defer func() {
		if err := os.Remove(gzipFileName); err != nil {
			log.Printf("Could not delete gzip file: %v\n", gzipFileName)
		}
		if err := os.Remove(strings.TrimSuffix(gzipFileName, ".gz")); err != nil {
			log.Printf("Could not delete decompressed file: %v\n", strings.TrimSuffix(gzipFileName, ".gz"))
		}
	}()

	scanner := bufio.NewScanner(outputFile) // ファイルを行ごとに読み込む
	for scanner.Scan() {
		line := scanner.Text()

		// 空のログは無視
		if line == "" {
			continue
		}

		var status_code string
		if c.attributes["logtype"] == "alb" {
			status_code = strings.Split(line, " ")[8]
		} else if c.attributes["logtype"] == "cloudfront" || c.attributes["logtype"] == "cdn" {
			if strings.Contains(line, "#Version:") || strings.Contains(line, "#Fields:") {
				continue // ヘッダー行は無視
			}
			status_code = strings.Split(line, "\t")[8]
		}

		// ログレベルに応じた送信
		switch {
		case regexp.MustCompile(`5\d{2}`).MatchString(status_code):
			c.logger.Error(line, "statuscode", status_code, "logtype", c.attributes["logtype"], "environment", c.attributes["environment"])
		case regexp.MustCompile(`4\d{2}`).MatchString(status_code):
			c.logger.Warn(line, "statuscode", status_code, "logtype", c.attributes["logtype"], "environment", c.attributes["environment"])
		case regexp.MustCompile(`[2-3]\d{2}`).MatchString(status_code):
			c.logger.Info(line, "statuscode", status_code, "logtype", c.attributes["logtype"], "environment", c.attributes["environment"])
		default:
			c.logger.Info(line, "statuscode", status_code, "logtype", c.attributes["logtype"], "environment", c.attributes["environment"])
		}
	}
	fmt.Printf("SendLog Goroutine finished (%s). runtime: %s\n", object, time.Since(startTime))
}

// ログのバッチを強制的に送信
func (c *OTLPClient) FlushLogs(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.loggerProvider.ForceFlush(ctx)
}

// OTLPクライアントを正常にシャットダウン
func (c *OTLPClient) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.loggerProvider.Shutdown(ctx)
}

func main() {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("ap-northeast-1"),
	)
	if err != nil {
		fmt.Println("Error creating config:", err)
		return
	}

	sqs_client := sqs.NewFromConfig(cfg)

	// キューからメッセージを受信するためのパラメータ
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: 10, // 一度に受信する最大メッセージ数
		VisibilityTimeout:   30, // メッセージが他の受信者から見えなくなる時間（秒）
		WaitTimeSeconds:     20, // ロングポーリングを有効にする時間（秒）
		// WaitTimeSeconds:     0, // ロングポーリングを無効にする
	}

	for {
		resp, err := sqs_client.ReceiveMessage(context.TODO(), receiveParams)
		if err != nil {
			fmt.Println("Error receiving message from SQS:", err)
			continue
		}

		if len(resp.Messages) == 0 {
			fmt.Println("Received no messages")
			continue
		}

		for _, message := range resp.Messages {
			var s3Event S3EventNotification
			if err := json.Unmarshal([]byte(*message.Body), &s3Event); err != nil {
				fmt.Println("Error unmarshalling message:", err)
				continue
			}

			for _, record := range s3Event.Records {
				logType := strings.Split(record.S3.Object.Key, "/")[0]
				sid := strings.Split((strings.Split(record.S3.Object.Key, "/")[1]), "-")[1]
				environment := strings.Split((strings.Split(record.S3.Object.Key, "/")[1]), "-")[0]

				attributes := map[string]string{
					"environment": environment,
					"logtype":     logType,
				}

				// OTLPクライアントの作成
				otlpClient, err := NewOTLPClient(context.Background(), sid, attributes)
				if err != nil {
					log.Fatalf("Failed to create OTLP client: %v", err)
				}

				go otlpClient.SendLog(cfg, record.S3.Bucket.Name, record.S3.Object.Key)
			}

			// 処理済みとしてメッセージを削除
			_, err := sqs_client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(QueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				fmt.Printf("Error deleting message: %v\n", err)
			}
		}

		// 次の確認まで待機
		time.Sleep(1 * time.Second)
	}
}
