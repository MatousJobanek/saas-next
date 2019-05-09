package cluster

import (
	"context"
	saasv1alpha1 "github.com/redhat-developer/saas-next/pkg/apis/saas/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

const (
	PlaneNamespaceName   = "saas-control-plane"
	ConfigClusterName    = "saas-next-clusterconfig"
	ServiceAccountName   = "saas-next"
	SaSecretMemberPrefix = "saas-next-member-"
)

var localConfig = LocalConfigHolder{}

type LocalConfigHolder struct {
	mux    sync.Mutex
	config *saasv1alpha1.ClusterConfig
}

func GetLocalClusterConfig(client client.Client) (*saasv1alpha1.ClusterConfig, error) {
	return localConfig.load(client)
}

func ResetLocalClusterConfig() {
	localConfig.mux.Lock()
	localConfig.config = nil
	localConfig.mux.Unlock()
}

func (h LocalConfigHolder) load(client client.Client) (*saasv1alpha1.ClusterConfig, error) {
	h.mux.Lock()
	defer h.mux.Unlock()
	if h.config != nil {
		return h.config, nil
	}
	clusterConfig := &saasv1alpha1.ClusterConfig{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: PlaneNamespaceName, Name: ConfigClusterName}, clusterConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return nil, err
	}
	h.config = clusterConfig
	return clusterConfig, nil
}
