package utilities

import (
	"log"
	"os"
	"path/filepath"
	_ "time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	repoEntry = &repo.Entry{
		Name: "opensearch",
		URL:  "https://opensearch-project.github.io/helm-charts/",
	}
)

func OpenSearchHelmSetting(releaseName string, actionType string) (*action.Install, *action.Uninstall, *chart.Chart) {
	// Helm CLI設定の取得
	settings := cli.New()
	// settings.Debug = true

	// Namespaceを設定
	settings.SetNamespace("opensearch")

	// リポジトリファイルのパスを設定
	repoFile := settings.RepositoryConfig

	// リポジトリファイルを読み込むか、存在しない場合は新規作成する
	var r *repo.File
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		r = repo.NewFile()
	} else {
		var err error
		r, err = repo.LoadFile(repoFile)
		if err != nil {
			log.Fatalf("Failed to load repo file: %v", err)
		}
	}

	// リポジトリファイルにリポジトリを追加する
	if !r.Has(repoEntry.Name) {
		r.Update(repoEntry)

		// リポジトリファイルを保存するディレクトリを作成
		if err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm); err != nil {
			log.Fatalf("Failed to create directory for repo file: %v", err)
		}

		// リポジトリファイルを保存する
		if err := r.WriteFile(repoFile, 0644); err != nil {
			log.Fatalf("Failed to write repo file: %v", err)
		}
	}

	// チャートリポジトリを作成し、インデックスファイルをダウンロードする
	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		log.Fatalf("Failed to create new chart repository: %v", err)
	}
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		log.Fatalf("Failed to download index file: %v", err)
	}

	// Helm設定の初期化
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), "opensearch", "secret", func(format string, v ...interface{}) {
		log.Printf(format, v...)
	}); err != nil {
		log.Fatalf("Failed to initialize Helm configuration: %v", err)
	}

	var installClient *action.Install
	var uninstallClient *action.Uninstall
	if actionType == "install" {
		installClient = action.NewInstall(actionConfig)
		// インストールクライアントの設定
		installClient.Namespace = "opensearch"
		installClient.ReleaseName = releaseName
		installClient.CreateNamespace = true
		// installClient.Wait = true ## k8sリソースがetcdに登録されるだけではなく、実際にrunning状態になるまで待つ。デフォルトはfalseでetcdに登録されるだけでプロンプトを返す
		// installClient.Timeout = 30 * time.Second ## Waitをtrueにした場合、どれくらい待つかを設定
		uninstallClient = nil
	} else if actionType == "uninstall" {
		uninstallClient = action.NewUninstall(actionConfig)
		installClient = nil
	}

	// チャートのパスを見つける
	chartName := "opensearch/opensearch"
	chartPathOptions := action.ChartPathOptions{}
	chartPath, err := chartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		log.Fatalf("Failed to locate chart: %v", err)
	}

	// チャートをロードする
	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("Failed to load chart: %v", err)
	}

	return installClient, uninstallClient, chart
}
