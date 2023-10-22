package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
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

func sendmsg(i int, c chan int, svc *sqs.SQS, wg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer wg.Done()
	defer sem.Release(1) // goroutineが完了したらリリース

	// 乱数生成器を初期化。これは一度だけ実行する必要がある。
	rand.Seed(time.Now().UnixNano())

	n := rand.Intn(1000000000000)

	MessageDedupId := strconv.Itoa(n)
	messageBody := "Hello, SQS from Go! " + strconv.Itoa(i)

	// メッセージの送信
	sendMsgInput := &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(QueueURL),
		// FIFOキューを使用している場合、MessageGroupIdが必要
		MessageGroupId: aws.String("SQS_TEST"), // 通常、同じ処理を行うメッセージに共通の値を設定
		// MessageDeduplicationIdはオプション
		MessageDeduplicationId: aws.String(MessageDedupId),
		// MessageDeduplicationId は、可能な限りユニークな値を提供することが重要（ただし、必須ではない）。
		// これは、同一の MessageDeduplicationId を持つメッセージが重複排除期間内に複数回送信された場合、後続のメッセージが受け入れられないことを意味する。(MessageDeduplicationIdが重複するメッセージはreceive側で受信しない)
	}
	_, err := svc.SendMessage(sendMsgInput)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	c <- i
}

func main() {
	c := make(chan int)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	svc := sqs.New(sess)

	sem := semaphore.NewWeighted(MaxConcurrentGoroutines) // セマフォを初期化
	ctx := context.TODO()                                 // 通常、キャンセルやタイムアウトが必要な場合には適切なコンテキストを使用

	var wg sync.WaitGroup

	count := 10

	for i := 1; i <= count; i++ {
		// セマフォを取得
		if err := sem.Acquire(ctx, 1); err != nil {
			fmt.Println("Failed to acquire semaphore:", err)
			continue
		}
		wg.Add(1)
		go sendmsg(i, c, svc, &wg, sem)
	}

	go func() {
		fmt.Println("Waiting for goroutines to finish")
		wg.Wait()
		close(c)
	}()

	for i := range c {
		fmt.Println("Message sent: ", i)
	}
}
