package utilities

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetKubeconfig() (*rest.Config, error) {
	// Kubernetesのクラスター設定を取得
	config, err := rest.InClusterConfig()
	// k8s pod内でないなら、エラーが返ってくる
	// その場合は次のローカルのkubeconfig取得処理へ
	if err == nil {
		return config, nil
	} else {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	// ホームディレクトリからkubeconfigのパスを取得
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// kubeconfigファイルを使用して設定をロード
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config from kubeconfig: %v\n", err)
		return nil, err
	}
	return config, nil
}
