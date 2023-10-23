// 受信を並列にすると同じメッセージが複数回処理されたので、並列化をやめてシングルスレッドで処理するようにした
package main

import (
	"context"
	"fmt"

	// "sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"golang.org/x/sync/semaphore"
)

const (
	QueueURL                = "https://sqs.ap-northeast-1.amazonaws.com/299413808364/lee_test.fifo"
	MaxConcurrentGoroutines = 2 // 同時に実行されるgoroutineの最大数
)

// var wg sync.WaitGroup
var sem = semaphore.NewWeighted(MaxConcurrentGoroutines) // セマフォを初期化

func recvmsg(message *sqs.Message, resp *sqs.ReceiveMessageOutput, svc *sqs.SQS, receiveParams *sqs.ReceiveMessageInput) {
	// defer wg.Done()
	// セマフォを取得
	if err := sem.Acquire(context.Background(), 1); err != nil { // 指定した同時実行数制限semから1つ実行権限を取得。上限に達していて取得できない場合は、取得でき次第、実行を開始
		fmt.Println("Failed to acquire semaphore:", err)
		return
	}
	defer sem.Release(1) // goroutineが完了したらリリース

	// 受信したメッセージを処理
	fmt.Printf("Message Body:  %s\n", *message.Body)
	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(QueueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		fmt.Println("Failed to delete message:", err)
	}
}

func main() {
	// AWSセッションを作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// SQSサービスのクライアントを作成
	svc := sqs.New(sess)

	// キューからメッセージを受信するためのパラメータ
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: aws.Int64(10), // 一度に受信する最大メッセージ数
		VisibilityTimeout:   aws.Int64(30), // メッセージが他の受信者から見えなくなる時間（秒）
		WaitTimeSeconds:     aws.Int64(20), // ロングポーリングを有効にする時間（秒）
		// WaitTimeSeconds: aws.Int64(0), // ロングポーリングを無効にする
	}

	for {
		// メッセージを受信
		resp, err := svc.ReceiveMessage(receiveParams)
		time.Sleep(1 * time.Second) // 1秒ごとに最大10件のメッセージを受信(処理)するためのsleep
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}

		// キューが空の場合、"Received no messages"を出力してループの最初に戻る
		if len(resp.Messages) == 0 {
			fmt.Println("Received no messages")
			continue
		}

		// 受信したメッセージを処理
		for _, message := range resp.Messages {
			// wg.Add(1)

			go recvmsg(message, resp, svc, receiveParams)
		}

		fmt.Println("----------------------------------------")
	}
}
