package saasuser

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	userv1 "github.com/openshift/api/user/v1"
	clientuserv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	errs "github.com/pkg/errors"
	saasv1alpha1 "github.com/redhat-developer/saas-next/pkg/apis/saas/v1alpha1"
	"github.com/redhat-developer/saas-next/pkg/cluster"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_saasuser")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SaasUser Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSaasUser{client: mgr.GetClient(), scheme: mgr.GetScheme(), config: mgr.GetConfig()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("saasuser-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SaasUser
	err = c.Watch(&source.Kind{Type: &saasv1alpha1.SaasUser{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	//err = c.Watch(&source.Kind{Type: &v1.User{}}, &handler.EnqueueRequestForObject{})
	//if err != nil {
	//	return err
	//}
	return nil
}

var _ reconcile.Reconciler = &ReconcileSaasUser{}

// ReconcileSaasUser reconciles a SaasUser object
type ReconcileSaasUser struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	config *rest.Config
}

// Reconcile reads that state of the cluster for a SaasUser object and makes changes based on the state read
// and what is in the SaasUser.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSaasUser) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SaasUser")

	// Fetch the SaasUser instance
	instance := &saasv1alpha1.SaasUser{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	if !instance.Spec.Approved {
		reqLogger.Info("the user is not approved - waiting for an approval to be able to provision it")
		return reconcile.Result{}, nil
	}

	config, err := cluster.GetLocalClusterConfig(r.client)
	if err != nil || config == nil {
		reqLogger.Error(err, "failed to get local cluster config")
		return reconcile.Result{}, err
	}

	if config.Spec.Config.ApiAddress == instance.Spec.TargetClusterAddress {
		err := r.createUser(reqLogger, instance)
		return reconcile.Result{}, err
	}

	if config.Spec.Config.Role == saasv1alpha1.Host {
		for _, member := range config.Spec.Config.Members {
			if member.ApiAddress == instance.Spec.TargetClusterAddress {

				memberClient, err := cluster.GetClusterClient(reqLogger, r.client, member)
				if err != nil {
					reqLogger.Error(err, "member client retrieval failed")
					return reconcile.Result{}, err
				}
				err = createSaasUser(instance, memberClient)
				if err != nil {
					reqLogger.Error(err, "failed creating saas user")
					return reconcile.Result{}, err
				}

				return reconcile.Result{}, nil
			}
		}
	}
	reqLogger.Error(err, "the target cluster wasn't found")
	return reconcile.Result{}, fmt.Errorf("the target cluster wasn't found")
}

func (r *ReconcileSaasUser) createUser(reqLogger logr.Logger, saasUser *saasv1alpha1.SaasUser) error {
	v1Client, err := clientuserv1.NewForConfig(r.config)
	if err != nil {
		reqLogger.Error(err, "failed to create local client")
		return err
	}
	user, err := v1Client.Users().Get(saasUser.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newUser := &userv1.User{
				ObjectMeta: v1.ObjectMeta{
					Name: saasUser.Name,
				},
				FullName: saasUser.Name,
			}
			_, err := v1Client.Users().Create(newUser)
			if err != nil {
				reqLogger.Error(err, "failed to create a new user")
				return err
			}
			reqLogger.WithValues("username", saasUser.Name).Info("user created")
			return nil
		}
		reqLogger.Error(err, "failed to get a user")
		return err
	}
	reqLogger.WithValues("username", user.Name).Info("user already exists")
	return nil
}

func createSaasUser(user *saasv1alpha1.SaasUser, client client.Client) error {
	u := &saasv1alpha1.SaasUser{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: user.Namespace, Name: user.Name}, u)
	if err != nil {
		if errors.IsNotFound(err) {
			u := &saasv1alpha1.SaasUser{
				ObjectMeta: v1.ObjectMeta{
					Name:      user.Name,
					Namespace: user.Namespace,
				},
				Spec: saasv1alpha1.SaasUserSpec{
					TargetClusterAddress: user.Spec.TargetClusterAddress,
				}}
			return client.Create(context.TODO(), u)
		}
		return errs.Wrapf(err, "failed to check if saasuser already exists")
	}
	return nil
}
