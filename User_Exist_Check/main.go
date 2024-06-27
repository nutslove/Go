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
	router *gin.Engine
	// 以下をグローバル変数として定義すると異なるリクエスト間で値が共有されてしまうため、関数内でローカル変数として定義すること！
	// no_db_user_count  int64
	// no_exist_db_user  []string
	// no_exist_iam_user []string
	// no_exist_os_user  []string
	wg               sync.WaitGroup
	mu               sync.Mutex
	PostgresUser     = os.Getenv("POSTGRES_USER")
	PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	PostgresDatabase = os.Getenv("POSTGRES_DB")
	dsn              = "port=5432 sslmode=disable TimeZone=Asia/Tokyo host=postgres" + " user=" + PostgresUser + " password=" + PostgresPassword + " dbname=" + PostgresDatabase
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

type UserExistCheck struct {
	Db_User       []string `json:"db_user"`
	Iam_User      []string `json:"iam_user"`
	Os_User       []string `json:"os_user"`
	SimpleAD_User []string `json:"simplead_user"`
	MsAD_User     []string `json:"ms_ad_user"`
}

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
	// _, ctx := router.Use(TracingMiddleware())

	// AWS SDKの設定をロード
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		fmt.Println("failed to load config")
		panic(err)
	}

	router.Use(AllowAllCORS())

	// router.GET("/", func(c *gin.Context) {
	router.POST("/", func(c *gin.Context) {
		// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")                                                                                                                            // すべてのオリジンを許可(CORS対策)
		// c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")                                                                                                     // 許可するHTTPメソッド
		// c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With") // 許可するHTTPヘッダー

		// if c.Request.Method == "OPTIONS" {
		// 	c.AbortWithStatus(http.StatusNoContent)
		// 	return
		// }

		// var user_type_count int // wg.Add()でいくつのgoroutineを作成するかを指定するために必要
		var request UserExistCheck
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("request:", request)
		fmt.Println("request.Db_User:", request.Db_User)
		fmt.Println("request.Iam_User:", request.Iam_User)
		fmt.Println("request.Os_User:", request.Os_User)
		fmt.Println("request.SimpleAD_User:", request.SimpleAD_User)
		fmt.Println("request.MsAD_User:", request.MsAD_User)

		tr := otel.Tracer("User-Exist-Check")
		ctx := c.Request.Context() // トレースのルートとなるコンテキストを生成
		// ctx, span := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
		// defer span.End()                               // トレースの終了
		ctx, parentSpan := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
		defer parentSpan.End()                               // トレースの終了

		c.Request = c.Request.WithContext(ctx) // コンテキストを更新
		c.Next()                               // 次のミドルウェアを呼び出し // ここでgin.Contextが更新される // この後の処理はgin.Contextの値を参照することができる

		// HTTPステータスコードが400以上の場合、エラーとしてマーク
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			// span.SetAttributes(attribute.Bool("error", true))
			parentSpan.SetAttributes(attribute.Bool("error", true))
		}

		// Add attributes to the span
		// span.SetAttributes(
		parentSpan.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.Int("http.status_code", statusCode),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.Request.RemoteAddr),
		)

		no_exist_db_user := []string{}
		no_exist_iam_user := []string{}
		no_exist_os_user := []string{}
		no_exist_simplead_user := []string{}
		no_exist_msad_user := []string{}

		// db_users := []string{"dbuser1", "dbuser2", "dbuser4", "dbuser5", "dbuser7"}
		if len(request.Db_User) != 0 {
			wg.Add(1)
			go DbuserExistCheck(ctx, &no_exist_db_user, request.Db_User...)
		}

		// iam_users := []string{"minorun365", "lee-testuser-for-iam", "iamuser3", "iamuser4", "iamuser5"}
		if len(request.Iam_User) != 0 {
			wg.Add(1)
			go AwsIamUserExistCheck(ctx, &no_exist_iam_user, cfg, request.Iam_User...)
		}

		// os_users := []string{"T232323", "Z121212", "Z343434", "M565656", "M090909", "M101010", "K232323"}
		if len(request.Os_User) != 0 {
			wg.Add(1)
			go OsUserExistCheck(ctx, &no_exist_os_user, cfg, request.Os_User...)
		}

		if len(request.SimpleAD_User) != 0 {
			wg.Add(1)
			go SimpleAdUserExistCheck(ctx, &no_exist_simplead_user, request.SimpleAD_User...)
		}

		if len(request.MsAD_User) != 0 {
			wg.Add(1)
			go MsAdUserExistCheck(ctx, &no_exist_msad_user, request.MsAD_User...)
		}

		wg.Wait() // wait until all goroutines are finished (including goroutines that are not created in this main function)

		c.JSON(http.StatusOK, gin.H{
			"no_exist_db_user":  no_exist_db_user,
			"no_exist_iam_user": no_exist_iam_user,
			"no_exist_os_user":  no_exist_os_user,
		})

		fmt.Println("no_exist_db_user:", no_exist_db_user)
		fmt.Println("no_exist_iam_user:", no_exist_iam_user)
		fmt.Println("no_exist_os_user:", no_exist_os_user)
	})

	router.Run(":8000") // listen and serve on
}

func DbuserExistCheck(ctx context.Context, no_exist_db_user *[]string, db_user ...string) {
	defer wg.Done()

	// トレーサーを取得
	tr := otel.Tracer("DbuserExistCheck")
	// 新しいSpanを開始
	_, span := tr.Start(ctx, "DbuserExistCheck")
	defer span.End()

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

	no_db_user_count := int64(0) // int64 型の数字 0 を代入
	for _, v := range db_user {
		db.Where("dbuser = ?", v).Find(&Dbuserpassword{}).Count(&no_db_user_count)
		if no_db_user_count == 0 {
			*no_exist_db_user = append(*no_exist_db_user, v)
			fmt.Printf("%v is not exist\n", v)
		} else {
			continue
		}
	}
}

func AwsIamUserExistCheck(ctx context.Context, no_exist_iam_user *[]string, cfg aws.Config, iam_user ...string) {
	defer wg.Done()

	// トレーサーを取得
	tr := otel.Tracer("AwsIamUserExistCheck")
	// 新しいSpanを開始
	_, span := tr.Start(ctx, "AwsIamUserExistCheck")
	defer span.End()

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
					*no_exist_iam_user = append(*no_exist_iam_user, v)
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

func OsUserExistCheck(ctx context.Context, no_exist_os_user *[]string, cfg aws.Config, os_user ...string) {
	defer wg.Done()

	// トレーサーを取得
	tr := otel.Tracer("OsUserExistCheck")
	// 新しいSpanを開始
	_, span := tr.Start(ctx, "OsUserExistCheck")
	defer span.End()

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

	f, err := os.Open("userlist")
	if err != nil {
		fmt.Println("failed to open file")
		panic(err)
	}
	defer f.Close()
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
				*no_exist_os_user = append(*no_exist_os_user, v)
				mu.Unlock()
			}
		}(v)
	}

	// fmt.Printf("read %d bytes\n", count)
	// fmt.Println("Data: ", string(data[:count]))
}

func SimpleAdUserExistCheck(ctx context.Context, no_exist_simplead_user *[]string, simplead_user ...string) {
	defer wg.Done()
	*no_exist_simplead_user = append(*no_exist_simplead_user, "simplead_user1")
}

func MsAdUserExistCheck(ctx context.Context, no_exist_msad_user *[]string, msad_user ...string) {
	defer wg.Done()
	*no_exist_msad_user = append(*no_exist_msad_user, "msad_user1")
}

func AllowAllCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")                                                                                                                            // すべてのオリジンを許可
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                                                                                             // 許可するHTTPメソッド
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With") // 許可するHTTPヘッダー

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// func TracingMiddleware() (gin.HandlerFunc, context.Context) {
// 	return func(c *gin.Context) {
// 		tr := otel.Tracer("User-Exist-Check")
// 		ctx := c.Request.Context() // トレースのルートとなるコンテキストを生成
// 		// ctx, span := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
// 		// defer span.End()                               // トレースの終了
// 		ctx, parentSpan := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
// 		defer parentSpan.End()                               // トレースの終了

// 		c.Request = c.Request.WithContext(ctx) // コンテキストを更新
// 		c.Next()                               // 次のミドルウェアを呼び出し // ここでgin.Contextが更新される // この後の処理はgin.Contextの値を参照することができる

// 		// HTTPステータスコードが400以上の場合、エラーとしてマーク
// 		statusCode := c.Writer.Status()
// 		if statusCode >= 400 {
// 			// span.SetAttributes(attribute.Bool("error", true))
// 			parentSpan.SetAttributes(attribute.Bool("error", true))
// 		}

// 		// Add attributes to the span
// 		// span.SetAttributes(
// 		parentSpan.SetAttributes(
// 			attribute.String("http.method", c.Request.Method),
// 			attribute.String("http.path", c.Request.URL.Path),
// 			attribute.String("http.host", c.Request.Host),
// 			attribute.Int("http.status_code", statusCode),
// 			attribute.String("http.user_agent", c.Request.UserAgent()),
// 			attribute.String("http.remote_addr", c.Request.RemoteAddr),
// 		)
// 	}, ctx
// }
