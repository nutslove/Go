// 受信を並列にすると同じメッセージが複数回処理されたので、並列化をやめてシングルスレッドで処理するようにした
package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	// "golang.org/x/sync/semaphore"
)

const (
	QueueURL = "https://sqs.ap-northeast-1.amazonaws.com/299413808364/lee_test.fifo"
	// MaxConcurrentGoroutines = 5 // 同時に実行されるgoroutineの最大数
)

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

	// sem := semaphore.NewWeighted(MaxConcurrentGoroutines) // セマフォを初期化
	// ctx := context.TODO()                                 // 通常、キャンセルやタイムアウトが必要な場合には適切なコンテキストを使用する

	// キューからメッセージを受信するためのパラメータ
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: aws.Int64(10), // 一度に受信する最大メッセージ数
		VisibilityTimeout:   aws.Int64(30), // メッセージが他の受信者から見えなくなる時間（秒）
		// WaitTimeSeconds:     aws.Int64(20), // ロングポーリングを有効にする時間（秒）
		WaitTimeSeconds: aws.Int64(0), // ロングポーリングを無効にする
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

		// var wg sync.WaitGroup

		// 受信したメッセージを処理
		for _, message := range resp.Messages {
			// wg.Add(1)

			// セマフォを利用してgoroutineの数を制限
			// if err := sem.Acquire(ctx, 1); err != nil {
			// 	wg.Done()
			// 	fmt.Println("Failed to acquire semaphore:", err)
			// 	continue
			// }

			fmt.Printf("Message Body:  %s\n", *message.Body)
			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(QueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				fmt.Println("Failed to delete message:", err)
			}

			// go func(msg *sqs.Message) {
			// 	defer wg.Done()
			// 	defer sem.Release(1) // goroutineが完了したらリリース

			// 	// メッセージの処理
			// 	// fmt.Printf("Message ID: %s\n", *message.MessageId)
			// 	fmt.Printf("Message Body: %s\n", *message.Body)

			// 	// メッセージの削除（削除しないとキューにメッセージが残り続ける）
			// 	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			// 		QueueUrl:      aws.String(QueueURL),
			// 		ReceiptHandle: message.ReceiptHandle,
			// 	})
			// 	if err != nil {
			// 		fmt.Println("Failed to delete message:", err)
			// 	}
			// }(message)
		}

		fmt.Println("----------------------------------------")
	}
}
