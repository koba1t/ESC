/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	escv1alpha2 "github.com/koba1t/ESC/api/v1alpha2"
)

// UserlandReconciler reconciles a Userland object
type UserlandReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=esc.k06.in,resources=userlands,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=esc.k06.in,resources=userlands/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=esc.k06.in,resources=templates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile loop for Userland resource
func (r *UserlandReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("userland", req.NamespacedName)

	// 1: Load the Userland resourcce by name
	var userland escv1alpha2.Userland
	if err := r.Get(ctx, req.NamespacedName, &userland); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Userland")
		} else {
			// Object not found, return.  Created objects are automatically garbage collected.
			log.Info("Userland object not found: " + req.NamespacedName.String())
		}
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2: Clean Up old Deployment which had been owned by Userland Resource.
	if err := r.cleanupOwnedResources(ctx, log, &userland); err != nil {
		log.Error(err, "failed to clean up old Deployment resources for this userland")
		return ctrl.Result{}, err
	}

	// 3: Create or Update deployment object
	templateName := userland.Spec.TemplateName
	// Get Template resource from templateName
	var template escv1alpha2.Template
	namespacedTemplateName := req.NamespacedName
	namespacedTemplateName.Name = templateName
	if err := r.Get(ctx, namespacedTemplateName, &template); err != nil {
		return ctrl.Result{}, err
	}

	// define deploymentName join to templateName and userland.Name
	deploymentName := userland.Spec.TemplateName + "-" + userland.Name

	// volume list of defined template resource
	volumes := []corev1.Volume{}

	for _, v := range template.Spec.VolumeSpecs {

		// pvc resource name
		pvcName := deploymentName + "-pvc-" + v.Name

		//add volume data for deploy.Spec.Template.Spec.Volumes
		vs := corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		}
		volumes = append(volumes, corev1.Volume{
			Name:         v.Name,
			VolumeSource: vs,
		})

		// define persistentVolumeClaim using deploymentName
		persistentVolumeClaim := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: req.Namespace,
			},
		}

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, persistentVolumeClaim, func() error {

			//get PersistentVolumeClaimSpec from template resource
			if persistentVolumeClaim.Spec.VolumeName != "" {
				v.PersistentVolumeClaimSpec.VolumeName = persistentVolumeClaim.Spec.VolumeName
			}
			persistentVolumeClaim.Spec = v.PersistentVolumeClaimSpec

			// set the owner so that garbage collection can kicks in
			if err := ctrl.SetControllerReference(&userland, persistentVolumeClaim, r.Scheme); err != nil {
				log.Error(err, "unable to set ownerReference from Userland to PersistentVolumeClaim")
				return err
			}

			return nil

		}); err != nil {
			// error handling of ctrl.CreateOrUpdate
			log.Error(err, "unable to ensure persistentVolumeClaim is correct")
			return ctrl.Result{}, err
		}
	}

	// define deployment template using deploymentName
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: req.Namespace,
		},
	}

	// Create or Update deployment object
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {

		// set the replicas from 1
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas

		//get template.Spec from template resource
		//templateSpec := templates.Items[0].Spec.Template.Spec
		templateSpec := template.Spec.Template.Spec

		// set a label for our deployment
		labels := map[string]string{
			"app":        deploymentName,
			"controller": req.Name,
			"template":   templateName,
		}

		// set labels to spec.selector for our deployment
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		// set containers to template.spec.containers for our deployment
		//if deploy.Spec.Template.Spec.Containers == nil {
		//	deploy.Spec.Template.Spec = templateSpec
		//}
		deploy.Spec.Template.Spec = templateSpec

		// set the owner so that garbage collection can kicks in
		if err := ctrl.SetControllerReference(&userland, deploy, r.Scheme); err != nil {
			log.Error(err, "unable to set ownerReference from Userland to Deployment")
			return err
		}

		if volumes != nil {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volumes...)
		}

		// end of ctrl.CreateOrUpdate
		return nil

	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err
	}

	// define service using deploymentName
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName + "-svc",
			Namespace: req.Namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {

		//get template.Spec from template resource
		//templateSpec := templates.Items[0].Spec.Template.Spec
		ServiceSpec := template.Spec.ServiceSpec

		// set a label for our deployment
		labels := map[string]string{
			"app":        deploymentName,
			"controller": req.Name,
			"template":   templateName,
		}

		//service.Spec = ServiceSpec
		// This code has error
		// spec.clusterIP: Invalid value: "": field is immutable
		if ServiceSpec.Ports != nil {
			service.Spec.Ports = ServiceSpec.Ports
		}

		// set labels to spec.selector for our deployment
		if service.Spec.Selector == nil {
			service.Spec.Selector = labels
		}

		//if service.Spec.Type == nil {
		//	service.Spec.Type = "ClusterIP"
		//}
		service.Spec.Type = "ClusterIP"

		// set the owner so that garbage collection can kicks in
		if err := ctrl.SetControllerReference(&userland, service, r.Scheme); err != nil {
			log.Error(err, "unable to set ownerReference from Userland to Service")
			return err
		}

		// end of ctrl.CreateOrUpdate
		return nil

	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure service is correct")
		return ctrl.Result{}, err
	}

	// 4: Update userland Status
	//TODO:

	return ctrl.Result{}, nil
}

// cleanupOwnedResources will delete any existing Deployment and Service resources
func (r *UserlandReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, userland *escv1alpha2.Userland) error {
	log.Info("finding existing Deployments for userland resource")

	// List all deployment resources owned by this Userland resource
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(userland.Namespace), client.MatchingFields(map[string]string{resourceOwnerKey: userland.Name})); err != nil {
		return err
	}

	// Delete deployment if the deployment name doesn't match userland.spec.TemplateName
	for _, deployment := range deployments.Items {
		if deployment.Name == userland.Spec.TemplateName+"-"+userland.Name {
			// If this deployment's name matches the one on the Userland resource
			// then do not delete it.
			continue
		}

		// Delete old deployment object which doesn't match userland.spec.TemplateName
		if err := r.Delete(ctx, &deployment); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}

		log.Info("delete deployment resource: " + deployment.Name)
		r.Recorder.Eventf(userland, corev1.EventTypeNormal, "Deleted", "Deleted deployment %q", deployment.Name)
	}

	// Service
	// List all Service resources owned by this Userland resource
	var services corev1.ServiceList
	if err := r.List(ctx, &services, client.InNamespace(userland.Namespace), client.MatchingFields(map[string]string{resourceOwnerKey: userland.Name})); err != nil {
		return err
	}

	// Delete deployment if the deployment name doesn't match userland.spec.TemplateName
	for _, service := range services.Items {
		if service.Name == userland.Spec.TemplateName+"-"+userland.Name+"-"+"svc" {
			// If this service's name matches the one on the Userland resource
			// then do not delete it.
			continue
		}

		// Delete old service object which doesn't match
		if err := r.Delete(ctx, &service); err != nil {
			log.Error(err, "failed to delete Service resource")
			return err
		}

		log.Info("delete service resource: " + service.Name)
		r.Recorder.Eventf(userland, corev1.EventTypeNormal, "Deleted", "Deleted service %q", service.Name)
	}

	// Service
	// List all Service resources owned by this Userland resource
	var persistentVolumeClaims corev1.PersistentVolumeClaimList
	if err := r.List(ctx, &persistentVolumeClaims, client.InNamespace(userland.Namespace), client.MatchingFields(map[string]string{resourceOwnerKey: userland.Name})); err != nil {
		return err
	}

	// // Delete deployment if the deployment name doesn't match userland.spec.TemplateName
	// for _, persistentVolumeClaim := range persistentVolumeClaims.Items {
	// 	if persistentVolumeClaim.Name == userland.Spec.TemplateName+"-"+userland.Name+"-"+"pvc" {
	// 		// If this service's name matches the one on the Userland resource
	// 		// then do not delete it.
	// 		continue
	// 	}

	// 	// Delete old service object which doesn't match
	// 	if err := r.Delete(ctx, &persistentVolumeClaim); err != nil {
	// 		log.Error(err, "failed to delete persistentVolumeClaim resource")
	// 		return err
	// 	}

	// 	log.Info("delete persistentVolumeClaim resource: " + persistentVolumeClaim.Name)
	// 	r.Recorder.Eventf(userland, corev1.EventTypeNormal, "Deleted", "Deleted persistentVolumeClaim %q", persistentVolumeClaim.Name)
	// }

	return nil
}

var (
	resourceOwnerKey = ".metadata.controller"
	apiGVStr         = escv1alpha2.GroupVersion.String()
)

// SetupWithManager setup with controller manager
func (r *UserlandReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// add resourceOwnerKey index to deployment object which Userland resource owns
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, resourceOwnerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Userland...
		if owner.APIVersion != apiGVStr || owner.Kind != "Userland" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// add resourceOwnerKey index to service object which Userland resource owns
	if err := mgr.GetFieldIndexer().IndexField(&corev1.Service{}, resourceOwnerKey, func(rawObj runtime.Object) []string {
		// grab the service object, extract the owner...
		service := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(service)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Userland...
		if owner.APIVersion != apiGVStr || owner.Kind != "Userland" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// add resourceOwnerKey index to persistentVolumeClaim object which Userland resource owns
	if err := mgr.GetFieldIndexer().IndexField(&corev1.PersistentVolumeClaim{}, resourceOwnerKey, func(rawObj runtime.Object) []string {
		// grab the service object, extract the owner...
		persistentVolumeClaim := rawObj.(*corev1.PersistentVolumeClaim)
		owner := metav1.GetControllerOf(persistentVolumeClaim)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Userland...
		if owner.APIVersion != apiGVStr || owner.Kind != "Userland" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// define to watch targets...Userland resource and owned Deployment
	return ctrl.NewControllerManagedBy(mgr).
		For(&escv1alpha2.Template{}).
		For(&escv1alpha2.Userland{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
