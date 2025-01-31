package pvc

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	kmccache "github.com/kyma-project/kyma-metrics-collector/pkg/cache"
	skrcommons "github.com/kyma-project/kyma-metrics-collector/pkg/skr/commons"
)

type Client struct {
	Resource  dynamic.NamespaceableResourceInterface
	ShootInfo kmccache.Record
}

func (c Config) NewClient(shootInfo kmccache.Record) (*Client, error) {
	restClientConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(shootInfo.KubeConfig))
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restClientConfig)
	if err != nil {
		return nil, err
	}

	nsResourceClient := dynamicClient.Resource(GroupVersionResource())

	return &Client{Resource: nsResourceClient, ShootInfo: shootInfo}, nil
}

func (c Client) List(ctx context.Context) (*corev1.PersistentVolumeClaimList, error) {
	unstructuredPVCList, err := c.Resource.Namespace(corev1.NamespaceAll).List(ctx, metaV1.ListOptions{})
	if err != nil {
		skrcommons.RecordSKRQuery(false, skrcommons.ListingPVCsAction, c.ShootInfo)
		return nil, err
	}

	skrcommons.RecordSKRQuery(true, skrcommons.ListingPVCsAction, c.ShootInfo)

	return convertUnstructuredListToPVCList(unstructuredPVCList)
}

func convertUnstructuredListToPVCList(unstructuredPVCList *unstructured.UnstructuredList) (*corev1.PersistentVolumeClaimList, error) {
	pvcList := new(corev1.PersistentVolumeClaimList)

	pvcListBytes, err := unstructuredPVCList.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pvcListBytes, pvcList)
	if err != nil {
		return nil, err
	}

	return pvcList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "persistentvolumeclaims",
	}
}
