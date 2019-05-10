package cluster

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	saasv1alpha1 "github.com/redhat-developer/saas-next/pkg/apis/saas/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

func GetClusterClient(log logr.Logger, cl client.Client, config saasv1alpha1.SaasClusterConfig, checkIfBound bool) (client.Client, error) {
	return registry.getClusterClient(log, cl, config, checkIfBound)
}

var registry = ClientRegistry{clients: map[string]client.Client{}}

type ClientRegistry struct {
	mux     sync.Mutex
	clients map[string]client.Client
}

func (r ClientRegistry) getClusterClient(log logr.Logger, cl client.Client, config saasv1alpha1.SaasClusterConfig, checkIfBound bool) (client.Client, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if checkIfBound && config.State == saasv1alpha1.Unbound {
		err := fmt.Errorf("cluster %s not bound", config.ApiAddress)
		log.Error(err, "the required cluster is not bound")
		return nil, err
	}

	if clusterClient, ok := r.clients[config.ApiAddress]; ok {
		return clusterClient, nil
	}

	secret := &v1.Secret{}
	namespacedSecretName := types.NamespacedName{Namespace: PlaneNamespaceName, Name: config.SecretRef.Name}
	err := cl.Get(context.TODO(), namespacedSecretName, secret)
	if err != nil {
		log.Error(err, "getting secret failed")
		return nil, err
	}
	token := secret.Data["token"]
	ca := secret.Data["ca.crt"]
	clusterConfig, err := clientcmd.BuildConfigFromFlags(config.ApiAddress, "")
	if err != nil {
		log.Error(err, "building config failed")
		return nil, err
	}
	clusterConfig.CAData = ca
	clusterConfig.BearerToken = string(token)

	clusterClient, err := client.New(clusterConfig, client.Options{})
	if err != nil {
		log.Error(err, "building cluster client failed")
		return nil, err
	}
	r.clients[config.ApiAddress] = clusterClient
	return clusterClient, nil
}
