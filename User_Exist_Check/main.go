package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "golang.org/x/sync/semaphore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_type "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3_type "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// const (
// 	MaxConcurrentGoroutines = 10 // 同時に実行されるgoroutineの最大数
// )

var (
	// sem               = semaphore.NewWeighted(MaxConcurrentGoroutines) // 同時に実行されるgoroutine数を管理(クローズするまでいつまでも実行可能)
	router            *gin.Engine
	no_db_user_count  int64
	no_exist_db_user  []string
	no_exist_iam_user []string
	no_exist_os_user  []string
	wg                sync.WaitGroup
	mu                sync.Mutex
	PostgresUser      = os.Getenv("POSTGRES_USER")
	PostgresPassword  = os.Getenv("POSTGRES_PASSWORD")
	PostgresDatabase  = os.Getenv("POSTGRES_DB")
	dsn               = "port=5432 sslmode=disable TimeZone=Asia/Tokyo host=postgres" + " user=" + PostgresUser + " password=" + PostgresPassword + " dbname=" + PostgresDatabase
)

type System struct {
	System string `gorm:"primaryKey;column:system"`
}

type Dbuserpassword struct {
	CombinationID int    `gorm:"primaryKey;column:combinationid"`
	System        string `gorm:"foreignKey:system;references:system;constraint:OnDelete:CASCADE;column:system"`
	Dbuser        string `gorm:"size:50;column:dbuser"`
	Password      string `gorm:"size:30;column:password"`
}

type Userprivilegestate struct {
	UserID         string
	CombinationID  int       `gorm:"foreignKey:combinationid;references:combinationid;constraint:OnDelete:CASCADE;column:combinationid"`
	StartTimeStamp time.Time `gorm:"type:timestamp with time zone;column:starttimestamp"`
	EndTimeStamp   time.Time `gorm:"type:timestamp with time zone;column:endtimestamp"`
	Index          int       `gorm:"primaryKey;autoIncrement;colum:index"`
}

// defaultではテーブル名は構造体名(末尾にsがつく)の複数形になるが、テーブル名を変更したい場合は、以下のようにTableName()を定義する
func (System) TableName() string {
	return "system"
}
func (Dbuserpassword) TableName() string {
	return "dbuserpassword"
}
func (Userprivilegestate) TableName() string {
	return "userprivilegestate"
}

func init() {
	// DB接続
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
		panic(err)
	}

	// テーブルの作成
	err = DB.AutoMigrate(&System{})
	if err != nil {
		fmt.Println("failed to migrate System database")
		panic(err)
	}

	err = DB.AutoMigrate(&Dbuserpassword{})
	if err != nil {
		fmt.Println("failed to migrate Dbuserpassword database")
		panic(err)
	}

	err = DB.AutoMigrate(&Userprivilegestate{})
	if err != nil {
		fmt.Println("failed to migrate Userprivilegestate database")
		panic(err)
	}
}

///////////// ToDo /////////////
// 【完】AWS SDKを使って、IAMユーザの存在チェックを実装すること！
// 【完】goroutineを使って並列処理を実装すること！
// 【完】テキストファイル(OSユーザリスト)をS3からダウンロードして読み込んでリストにOSユーザが存在するか実装！
// Ginを使ってGetリクエストを受け取って、存在しないAD,LDAP,IAM,DB,OSユーザの情報をjson形式で返すように実装！
// → ADとLDAP以外は実装完了。DBは実態に合わせて実装すること
// json形式でうけとったリクエストを構造体に変換すること！
// gin-with-otel側でこのuser-exist-checkをGETリクエストで呼び出して、gin-with-otel側でuser-exist-checkからの結果をjson形式で受け取って表示するように実装！
// IAM,DB,OS,ADなど処理ごとにSpanを作成すること！
// エラーハンドリングを実装すること！（どういう時にpanicするか、どういう時にエラーを返すか、どういう時にログを出力するかなど）
// jsonを扱う練習
// テストコードを書くこと！
// ログ出力を実装すること！
// OpenTelemetry設定を追加すること！
// context.TODO()が何か調べること！
// AWS GO SDK V2のの仕様ｋを調べること！(config.LoadDefaultConfigやNewFromConfig、errors.Asやsmithy.APIErrorなど)
///////////// ToDo /////////////

func main() {
	// Ginの設定
	router = gin.Default()

	// OTLPエクスポーターの設定
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithInsecure(), // TLSを無効にする場合に指定
	)
	if err != nil {
		fmt.Printf("Failed to create exporter: %v", err)
		panic(err)
	}

	// Tracerの設定
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL, // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
			semconv.ServiceNameKey.String("User-Exist-Check"), // ServiceNameKey is the key used to identify the service name in a Resource.
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	router = gin.Default()
	router.Use(TracingMiddleware())

	// AWS SDKの設定をロード
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		fmt.Println("failed to load config")
		panic(err)
	}

	router.GET("/", func(c *gin.Context) {
		wg.Add(3)

		db_users := []string{"dbuser1", "dbuser2", "dbuser4", "dbuser5", "dbuser7"}
		go DbuserExistCheck(db_users...)

		iam_users := []string{"minorun365", "lee-testuser-for-iam", "iamuser3", "iamuser4", "iamuser5"}
		go AwsIamUserExistCheck(cfg, iam_users...)

		os_users := []string{"T232323", "Z121212", "Z343434", "M565656", "M090909", "M101010", "K232323"}
		go OsUserExistCheck(cfg, os_users...)

		wg.Wait() // wait until all goroutines are finished (including goroutines that are not created in this main function)

		c.JSON(http.StatusOK, gin.H{
			"no_exist_db_user":  no_exist_db_user,
			"no_exist_iam_user": no_exist_iam_user,
			"no_exist_os_user":  no_exist_os_user,
		})

		fmt.Println("no_exist_db_user:", no_exist_db_user)
		fmt.Println("no_exist_iam_user:", no_exist_iam_user)
		fmt.Println("no_exist_os_user:", no_exist_os_user)

		// 初期化 (スライスの中身を空にする) しないと次のリクエストで前回の結果が残ってしまう
		no_exist_db_user = nil
		no_exist_iam_user = nil
		no_exist_os_user = nil
	})

	router.Run(":8000") // listen and serve on
}

func DbuserExistCheck(db_user ...string) {
	defer wg.Done()

	// DB接続
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
		panic(err)
	}

	// // セッションを利用するために、DBに接続する
	// sqldb, err := db.DB()
	// if err != nil {
	// 	fmt.Println("failed to connect DB")
	// 	panic(err)
	// }
	// // DBに対する処理が終わったら、DB接続を解除する
	// defer sqldb.Close()

	for _, v := range db_user {
		db.Where("dbuser = ?", v).Find(&Dbuserpassword{}).Count(&no_db_user_count)
		if no_db_user_count == 0 {
			no_exist_db_user = append(no_exist_db_user, v)
			fmt.Printf("%v is not exist\n", v)
		} else {
			continue
		}
	}
}

func AwsIamUserExistCheck(cfg aws.Config, iam_user ...string) {
	defer wg.Done()

	svc := iam.NewFromConfig(cfg)

	// IAMユーザの存在チェック
	wg.Add(len(iam_user))
	for _, v := range iam_user {
		go func(v string) {
			defer wg.Done()

			input := &iam.GetUserInput{
				UserName: aws.String(v),
			}

			_, err := svc.GetUser(context.TODO(), input)
			if err != nil {
				var nsk *iam_type.NoSuchEntityException
				if errors.As(err, &nsk) {
					fmt.Println("NoSuchEntityException")
					mu.Lock()
					no_exist_iam_user = append(no_exist_iam_user, v)
					mu.Unlock()
				}
				var apiErr smithy.APIError
				if errors.As(err, &apiErr) {
					fmt.Println("StatusCode:", apiErr.ErrorCode(), ", Msg:", apiErr.ErrorMessage())
				}
			}
		}(v)
	}
	// wg.Wait()
	// wg.Wait()はすべてのgoroutineが完了するまで待つため、ここにwg.Wait()があるとmain関数内のgoroutineが完了するまで待つことになり、
	// main関数内のwg.Wait()とここにあるwg.Wait()の2つのwg.Wait()がお互いを待ち合うことになり、デッドロックが発生する。
}

func OsUserExistCheck(cfg aws.Config, os_user ...string) {
	defer wg.Done()

	svc := s3.NewFromConfig(cfg)

	// S3バケットからオブジェクトのコンテンツを含むHTTPレスポンスのボディを取得 (これで直接ファイルとしてダウンロードするわけではない)
	objectName := "userlist"
	input := &s3.GetObjectInput{
		Bucket: aws.String("lee-for-user-exist-check-test"),
		Key:    aws.String(objectName),
	}

	result, err := svc.GetObject(context.TODO(), input)
	if err != nil {
		var nsk *s3_type.NoSuchKey
		if errors.As(err, &nsk) {
			fmt.Println("NoSuchKey")
		}

		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			fmt.Println("StatusCode:", apiErr.ErrorCode(), ", Msg:", apiErr.ErrorMessage())
		}
		return
	}

	// レスポンスのボディをファイルに書き込む
	defer result.Body.Close()
	file, err := os.Create(objectName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", objectName, err)
		panic(err)
	}
	defer file.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectName, err)
		panic(err)
	}
	_, err = file.Write(body)
	if err != nil {
		log.Printf("Couldn't write object body to file %v. Here's why: %v\n", objectName, err)
		panic(err)
	}

	// // ディレクトリの内容を読み込む
	// files, err := os.ReadDir(".")
	// if err != nil {
	// 	fmt.Printf("Failed to read directory: %v\n", err)
	// }
	// // ファイルとディレクトリの名前を表示
	// for _, file := range files {
	// 	fmt.Println(file.Name())
	// }

	// data := make([]byte, 1024) // 1024byteのスライスを作成. 1024byteより大きいデータがある場合は動的に拡張される
	f, err := os.Open("userlist")
	defer f.Close()
	if err != nil {
		fmt.Println("failed to open file")
		panic(err)
	}
	// count, err := f.Read(data)
	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("failed to read file")
		panic(err)
	}

	// OSユーザの存在チェック
	wg.Add(len(os_user))
	for _, v := range os_user {
		// ファイルの中身から該当os_userが存在するか確認
		go func(v string) {
			defer wg.Done()
			if strings.Contains(string(data), v) {
				fmt.Printf("String '%s' found in '%s'\n", v, objectName)
			} else {
				fmt.Printf("String '%s' not found in file '%s'\n", v, objectName)
				mu.Lock()
				no_exist_os_user = append(no_exist_os_user, v)
				mu.Unlock()
			}
		}(v)
	}

	// fmt.Printf("read %d bytes\n", count)
	// fmt.Println("Data: ", string(data[:count]))
}

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tr := otel.Tracer("User-Exist-Check")
		ctx := c.Request.Context()                     // トレースのルートとなるコンテキストを生成
		ctx, span := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
		defer span.End()                               // トレースの終了

		c.Request = c.Request.WithContext(ctx) // コンテキストを更新
		c.Next()                               // 次のミドルウェアを呼び出し // ここでgin.Contextが更新される // この後の処理はgin.Contextの値を参照することができる

		// HTTPステータスコードが400以上の場合、エラーとしてマーク
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		// Add attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.Int("http.status_code", statusCode),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.Request.RemoteAddr),
		)
	}
}
