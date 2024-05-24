package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func CreateCaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	caas_id := c.Param("caas_id")

	// Tracerの設定
	tr := otel.Tracer("Create CaaS Cluster")
	_, span := tr.Start(ctx, "Create Namespace", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// Namespaceを作成するマニフェストの定義
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: caas_id,
			Labels: map[string]string{
				"target-namespace": "metrics",
				"app":              "caas",
			},
		},
	}

	// Namespaceが存在するか確認、Namespace作成 (指定したnamespaceがすでに存在する場合はerrはnilになる)
	// Namespaceが存在する場合は以降の処理をスキップ
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace.Name, metav1.GetOptions{})
	if err != nil {
		_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Error creating namespace: %v\n", err)
		}
		fmt.Printf("Namespace[%v] created successfully\n", namespace.Name)
		span.End()
	} else {
		fmt.Printf("Namespace already exists: %v\n", ns)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("%s namespace already exists", caas_id),
		})
		span.End()
		return
	}

	_, span2 := tr.Start(ctx, "Create ResourceQuota", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// ResourceQuotaを作成するマニフェストの定義
	resourceQuota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("quota-%s", caas_id),
			Namespace: caas_id,
			Labels: map[string]string{
				"app": "caas",
			},
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceRequestsCPU:    resource.MustParse("10"),
				v1.ResourceRequestsMemory: resource.MustParse("10Gi"),
				v1.ResourcePods:           resource.MustParse("20"),
			},
		},
	}

	// ResourceQuotaを作成（ResourceQuotas()内のパラメータはnamespaceを指しており、必須）
	_, err = clientset.CoreV1().ResourceQuotas(caas_id).Create(context.TODO(), resourceQuota, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating resourcequota: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error creating resourcequota for %s\n Error messages: %s", caas_id, err),
		})
		span2.End()
		return
	} else {
		span2.End()
	}

	_, span3 := tr.Start(ctx, "Create LimitRange", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// LimitRangeを作成するマニフェストの定義
	limitRange := &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("limit-%s", caas_id),
			Namespace: caas_id,
			Labels: map[string]string{
				"app": "caas",
			},
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypePod,
					Max: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("4000m"),
						v1.ResourceMemory: resource.MustParse("2048Mi"),
					},
					MaxLimitRequestRatio: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2.1"),
						v1.ResourceMemory: resource.MustParse("2.1"),
					},
				},
			},
		},
	}

	// LimitRangeを作成
	_, err = clientset.CoreV1().LimitRanges(caas_id).Create(context.TODO(), limitRange, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating limitrange: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error creating limitrange for %s\n Error messages: %s", caas_id, err),
		})
		span3.End()
		return
	} else {
		span3.End()
	}

	_, span4 := tr.Start(ctx, "Create RoleBinding", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// RoleBindingを作成するマニフェストの定義
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("cass-user-role-%s", caas_id),
			Namespace: caas_id,
			Labels: map[string]string{
				"app": "caas",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     caas_id,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "caas-tenant-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// RoleBindingを作成
	_, err = clientset.RbacV1().RoleBindings(caas_id).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating rolebinding: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error creating rolebinding for %s\n Error messages: %s", caas_id, err),
		})
		span4.End()
		return
	} else {
		span4.End()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Created CaaS for %s successfully", caas_id),
	})
}

func GetCaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	caas_id := c.Param("caas_id")

	// Traceの設定
	tr := otel.Tracer("Get CaaS Cluster")
	_, span := tr.Start(ctx, "Get Namespace", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// Namespaceを取得
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), caas_id, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting namespace: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error getting namespace for %s\n Error messages: %s", caas_id, err),
		})
		span.End()
		return
	} else {
		span.End()
	}

	_, span2 := tr.Start(ctx, "Get ResourceQuota", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// ResourceQuotaを取得
	resourcequota, err := clientset.CoreV1().ResourceQuotas(caas_id).Get(context.TODO(), fmt.Sprintf("quota-%s", caas_id), metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting resourcequota: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error getting resourcequota for %s\n Error messages: %s", caas_id, err),
		})
		span2.End()
		return
	} else {
		span2.End()
	}

	_, span3 := tr.Start(ctx, "Get LimitRange", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// LimitRangeを取得
	limitrange, err := clientset.CoreV1().LimitRanges(caas_id).Get(context.TODO(), fmt.Sprintf("limit-%s", caas_id), metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting limitrange: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error getting limitrange for %s\n Error messages: %s", caas_id, err),
		})
		span3.End()
		return
	} else {
		span3.End()
	}

	_, span4 := tr.Start(ctx, "Get RoleBinding", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// RoleBindingを取得
	rolebinding, err := clientset.RbacV1().RoleBindings(caas_id).Get(context.TODO(), fmt.Sprintf("cass-user-role-%s", caas_id), metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting rolebinding: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error getting rolebinding for %s\n Error messages: %s", caas_id, err),
		})
		span4.End()
		return
	} else {
		span4.End()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": gin.H{
			"Namespace":     ns,
			"ResourceQuota": resourcequota,
			"LimitRange":    limitrange,
			"RoleBinding":   rolebinding,
		},
	})
}

func DeleteCaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	caas_id := c.Param("caas_id")

	// Traceの設定
	tr := otel.Tracer("Delete CaaS Cluster")

	_, span := tr.Start(ctx, "Delete ResourceQuota", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// ResourceQuotaを削除
	err := clientset.CoreV1().ResourceQuotas(caas_id).Delete(context.TODO(), fmt.Sprintf("quota-%s", caas_id), metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error deleting resourcequota: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error deleting resourcequota for %s\n Error messages: %s", caas_id, err),
		})
		span.End()
		return
	} else {
		span.End()
	}

	_, span2 := tr.Start(ctx, "Delete LimitRange", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// LimitRangeを削除
	err = clientset.CoreV1().LimitRanges(caas_id).Delete(context.TODO(), fmt.Sprintf("limit-%s", caas_id), metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error deleting limitrange: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error deleting limitrange for %s\n Error messages: %s", caas_id, err),
		})
		span2.End()
		return
	} else {
		span2.End()
	}

	_, span3 := tr.Start(ctx, "Delete RoleBinding", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", caas_id)))

	// RoleBindingを削除
	err = clientset.RbacV1().RoleBindings(caas_id).Delete(context.TODO(), fmt.Sprintf("cass-user-role-%s", caas_id), metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error deleting rolebinding: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error deleting rolebinding for %s\n Error messages: %s", caas_id, err),
		})
		span3.End()
		return
	} else {
		span3.End()
	}

	_, span4 := tr.Start(ctx, "Delete Namespace", trace.WithAttributes(attribute.String("service.name", "CaaS"), attribute.String("tenant", "caas_id")))

	// Namespaceを削除
	err = clientset.CoreV1().Namespaces().Delete(context.TODO(), caas_id, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error deleting namespace: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("Error deleting namespace for %s\n Error messages: %s", caas_id, err),
		})
		span4.End()
		return
	} else {
		span4.End()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Deleted CaaS for %s successfully", caas_id),
	})
}

func GetCaases(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get caases",
	})
}
