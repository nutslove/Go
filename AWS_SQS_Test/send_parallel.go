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

var wg sync.WaitGroup
var sem = semaphore.NewWeighted(MaxConcurrentGoroutines) // セマフォを初期化

func sendmsg(i int, svc *sqs.SQS) {
	defer wg.Done()
	// セマフォを取得
	if err := sem.Acquire(context.Background(), 1); err != nil { // 指定した同時実行数制限semから1つ実行権限を取得。上限に達していて取得できない場合は、取得でき次第、実行を開始
		fmt.Println("Failed to acquire semaphore:", err)
		return
	}
	defer sem.Release(1) // goroutineが完了したらリリース

	n := rand.Intn(100000000000000)

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
	fmt.Println("Message sent:  Hello, SQS from Go!", i)
}

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	svc := sqs.New(sess)

	// 乱数生成器を初期化。これは一度だけ実行する必要がある。
	rand.Seed(time.Now().UnixNano())

	count := 10

	for i := 1; i <= count; i++ {
		wg.Add(1)
		go sendmsg(i, svc)
	}
	wg.Wait()
}
