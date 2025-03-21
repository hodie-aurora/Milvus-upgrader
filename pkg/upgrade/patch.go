package upgrade

import (
	"context"
	"fmt"

	"github.com/hodie-aurora/milvus-upgrader/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PatchUpgrade performs a patch-level upgrade
func PatchUpgrade(client *k8s.ClientSet, instance, namespace, targetVersion string) error {
	cr, err := getMilvusCR(client, namespace, instance)
	if err != nil {
		return fmt.Errorf("failed to get Milvus CR: %v", err)
	}
	unstructuredCr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	if err != nil {
		return fmt.Errorf("failed to convert CR to unstructured: %v", err)
	}
	components, found, err := unstructured.NestedMap(unstructuredCr, "spec", "components")
	if !found || err != nil {
		return fmt.Errorf("failed to get spec.components: %v", err)
	}
	components["image"] = "milvusdb/milvus:" + targetVersion
	err = unstructured.SetNestedMap(unstructuredCr, components, "spec", "components")
	if err != nil {
		return fmt.Errorf("failed to set spec.components.image: %v", err)
	}
	gvr := schema.GroupVersionResource{
		Group:    "milvus.io",
		Version:  "v1beta1",
		Resource: "milvuses",
	}
	_, err = client.DynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), &unstructured.Unstructured{Object: unstructuredCr}, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Milvus CR: %v", err)
	}
	fmt.Println("Patch-level upgrade started, please check cluster status.")
	return nil
}
