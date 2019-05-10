package clusterconfig

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	saasv1alpha1 "github.com/redhat-developer/saas-next/pkg/apis/saas/v1alpha1"
	"github.com/redhat-developer/saas-next/pkg/cluster"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clusterconfig")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ClusterConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterConfig
	err = c.Watch(&source.Kind{Type: &saasv1alpha1.ClusterConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterConfig{}

// ReconcileClusterConfig reconciles a ClusterConfig object
type ReconcileClusterConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ClusterConfig object and makes changes based on the state read
// and what is in the ClusterConfig.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ClusterConfig")
	// Fetch the ClusterConfig config
	cluster.ResetLocalClusterConfig()
	config, err := cluster.GetLocalClusterConfig(r.client)
	if err != nil {
		return reconcile.Result{}, err
	} else if config == nil {
		return reconcile.Result{}, nil
	}

	if config.Spec.Config.Role == saasv1alpha1.Member {
		return r.ReconcileMember(reqLogger, request, config)
	} else {
		return r.ReconcileHost(reqLogger, request, config)
	}
}

func (r *ReconcileClusterConfig) ReconcileHost(log logr.Logger, request reconcile.Request, config *saasv1alpha1.ClusterConfig) (reconcile.Result, error) {
	for i := 0; i < len(config.Spec.Config.Members); i++ {
		member := config.Spec.Config.Members[i]
		if member.State == saasv1alpha1.Unbound || member.State == "" {
			memberClient, err := cluster.GetClusterClient(log, r.client, member, false)
			if err != nil {
				log.Error(err, "creating member client failed")
				return reconcile.Result{}, nil
			}
			memberConfig := &saasv1alpha1.ClusterConfig{}
			configNsdName := types.NamespacedName{Namespace: cluster.PlaneNamespaceName, Name: cluster.ConfigClusterName}
			err = memberClient.Get(context.TODO(), configNsdName, memberConfig)
			if err != nil {
				log.Error(err, "geting cluster config from member failed")
				return reconcile.Result{}, nil
			}
			if memberConfig.Spec.Config.Host.ApiAddress == config.Spec.Config.ApiAddress {
				memberConfig.Spec.Config.Host.State = saasv1alpha1.BoundAsynchronized
				err := memberClient.Update(context.TODO(), memberConfig)
				if err != nil {
					log.Error(err, "updating cluster config in member failed")
					return reconcile.Result{}, nil
				}
				config.Spec.Config.Members[i].State = saasv1alpha1.Bound
				err = r.client.Update(context.TODO(), config)
				if err != nil {
					log.Error(err, "updating local cluster config in host failed")
					return reconcile.Result{}, nil
				}
			} else {
				log.Error(fmt.Errorf("wrong host %s", memberConfig.Spec.Config.Host.ApiAddress), "the host in member doesn't match")
				return reconcile.Result{}, nil
			}
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileClusterConfig) ReconcileMember(log logr.Logger, request reconcile.Request, config *saasv1alpha1.ClusterConfig) (reconcile.Result, error) {
	if config.Spec.Config.Host.State == saasv1alpha1.Bound {
		return reconcile.Result{}, nil
	}
	if config.Spec.Config.Host.State == saasv1alpha1.BoundAsynchronized {
		return r.SynchronizeMember(log, request, config)
	}

	host := config.Spec.Config.Host
	hostClient, err := cluster.GetClusterClient(log, r.client, host, false)
	if err != nil {
		log.Error(err, "creating host client failed")
		return reconcile.Result{}, nil
	}

	hostConfig := &saasv1alpha1.ClusterConfig{}
	err = hostClient.Get(context.TODO(), request.NamespacedName, hostConfig)
	if err != nil {
		log.Error(err, "getting host config failed")
		return reconcile.Result{}, nil
	}
	isPresent := false
	for _, member := range hostConfig.Spec.Config.Members {
		if member.ApiAddress == config.Spec.Config.ApiAddress {
			isPresent = true
			break
		}
	}
	if !isPresent {
		sa := &v1.ServiceAccount{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cluster.PlaneNamespaceName, Name: cluster.ServiceAccountName}, sa)
		if err != nil {
			log.Error(err, "getting local SA failed")
			return reconcile.Result{}, nil
		}

		saTokenList := &v1.SecretList{}
		err = r.client.List(context.TODO(), &client.ListOptions{Namespace: cluster.PlaneNamespaceName}, saTokenList)
		if err != nil {
			log.Error(err, "getting local SA secret list failed")
			return reconcile.Result{}, nil
		}
		var saToken *v1.Secret
		for _, secret := range saTokenList.Items {
			if isOwnedBy(secret, sa) && secret.Type == v1.SecretTypeServiceAccountToken {
				saToken = &secret
			}
		}
		if saToken == nil {
			log.Error(err, "local SA secret not found")
			return reconcile.Result{}, nil
		}

		tokenToCreate := &v1.Secret{
			Data:       saToken.Data,
			StringData: saToken.StringData,
		}
		tokenToCreate.Namespace = cluster.PlaneNamespaceName
		tokenToCreate.GenerateName = cluster.SaSecretMemberPrefix

		err = hostClient.Create(context.TODO(), tokenToCreate)
		if err != nil {
			log.Error(err, "creating member SA secret in host failed")
			return reconcile.Result{}, nil
		}

		hostConfig.Spec.Config.Members = append(hostConfig.Spec.Config.Members,
			saasv1alpha1.SaasClusterConfig{
				ApiAddress: config.Spec.Config.ApiAddress,
				SecretRef: &saasv1alpha1.SecretRef{
					Name: tokenToCreate.Name,
				},
				State: saasv1alpha1.Unbound,
			})

		err = hostClient.Update(context.TODO(), hostConfig)
		if err != nil {
			log.Error(err, "updating host config failed")
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileClusterConfig) SynchronizeMember(log logr.Logger, request reconcile.Request, config *saasv1alpha1.ClusterConfig) (reconcile.Result, error) {
	// todo synchronize all CRs
	config.Spec.Config.Host.State = saasv1alpha1.Bound
	err := r.client.Update(context.TODO(), config)
	if err != nil {
		log.Error(err, "updating local member config failed")
	}
	return reconcile.Result{}, nil
}

func isOwnedBy(secret v1.Secret, sa *v1.ServiceAccount) bool {
	for _, sec := range sa.Secrets {
		if sec.Name == secret.Name {
			return true
		}
	}
	return false
}
