package saasnamespace

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	saasv1alpha1 "github.com/redhat-developer/saas-next/pkg/apis/saas/v1alpha1"
	"github.com/redhat-developer/saas-next/pkg/cluster"
	"github.com/redhat-developer/saas-next/pkg/controller/clusterconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

var log = logf.Log.WithName("controller_saasnamespace")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SaasNamespace Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSaasNamespace{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("saasnamespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SaasNamespace
	err = c.Watch(&source.Kind{Type: &saasv1alpha1.SaasNamespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SaasNamespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSaasNamespace{}

// ReconcileSaasNamespace reconciles a SaasNamespace object
type ReconcileSaasNamespace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SaasNamespace object and makes changes based on the state read
// and what is in the SaasNamespace.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSaasNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Namespace")

	// Fetch the Namespace instance
	instance := &corev1.Namespace{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			// todo handle removed
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	targetClient, err := r.getTargetClient(reqLogger)
	if err != nil {
		return reconcile.Result{}, nil
	}

	userName, ok := instance.Annotations["openshift.io/requester"]
	if !ok {
		reqLogger.Info("there is no \"openshift.io/requester\" label")
		return reconcile.Result{}, nil
	}

	saasUser := &saasv1alpha1.SaasUser{}
	err = targetClient.Get(context.TODO(), types.NamespacedName{Namespace: cluster.PlaneNamespaceName, Name: userName}, saasUser)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("the namespace owner is not saas user")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "failed to get saas user")
		return reconcile.Result{}, nil
	}
	err = r.createSaasNamespace(reqLogger, targetClient, userName, request.Name)
	return reconcile.Result{}, err
}

func (r *ReconcileSaasNamespace) createSaasNamespace(reqLogger logr.Logger, targetClient client.Client, userName, namespace string) error {
	// todo solve conflicts
	name := fmt.Sprintf("%s-%s", userName, namespace)

	saasNs := &saasv1alpha1.SaasNamespace{}
	namespacedName := types.NamespacedName{Namespace: cluster.PlaneNamespaceName, Name: name}
	err := targetClient.Get(context.TODO(), namespacedName, saasNs)
	if err != nil {
		if errors.IsNotFound(err) {
			saasNs.Spec.NamespaceName = namespace
			saasNs.Spec.Owner = userName
			saasNs.Namespace = cluster.PlaneNamespaceName
			saasNs.ObjectMeta.Name = name
			err := targetClient.Create(context.TODO(), saasNs)
			if err != nil {
				reqLogger.Error(err, "failed to create saas namespace")
				return err
			}
			return err
		}
		reqLogger.Error(err, "failed to get saas namespace")
		return err
	}
	return nil
}

func (r *ReconcileSaasNamespace) getTargetClient(reqLogger logr.Logger) (client.Client, error) {
	config, err := clusterconfig.GetLocalClusterConfig(r.client)
	if err != nil {
		reqLogger.Error(err, "failed to get local cluster config")
		return nil, err
	}
	if config == nil {
		reqLogger.Error(err, "local cluster config not found")
		return nil, fmt.Errorf("local cluster config not found")
	}

	targetClient := r.client
	if config.Spec.Config.Role != saasv1alpha1.Host {
		targetClient, err = cluster.GetClusterClient(reqLogger, r.client, config.Spec.Config.Host)
		if err != nil {
			reqLogger.Error(err, "failed to create host client")
			return nil, err
		}
	}
	return targetClient, nil
}
