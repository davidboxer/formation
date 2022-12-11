package controller

import (
	"context"
	errors2 "errors"
	"github.com/Doout/formation/internal/utils"
	"github.com/Doout/formation/types"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"github.com/snorwin/jsonpatch"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"
	"unsafe"
)

type Controller struct {
	cli    client.Client
	scheme *runtime.Scheme
	object client.Object
}

func NewController(scheme *runtime.Scheme, cli client.Client) *Controller {
	return &Controller{scheme: scheme, cli: cli}
}

func (c Controller) ForObject(object client.Object) *Controller {
	b := &Controller{cli: c.cli, scheme: c.scheme, object: object}
	return b
}

func (c Controller) Reconcile(ctx context.Context, list []types.Resource) (result ctrl.Result, err error) {
	status, err := c.GetStatus()
	if err != nil {
		return ctrl.Result{}, err
	}

	//Build status map
	statusMap := map[string]*types.ResourceStatus{}
	for idx, res := range status.Resources {
		if res == nil {
			continue
		}
		key := strings.ToLower(res.Type + "/" + res.Name)
		statusMap[key] = status.Resources[idx]
	}

	resourceMap := map[string]types.Resource{}
	//Go over each resource and check if it exists in the status, if not, add.
	//This task need to be done every reconcile as this list might be outdated on the next call.
	var patch client.Patch
	for idx, res := range list {
		key := strings.ToLower(res.Type() + "/" + res.Name())
		resourceMap[key] = list[idx]
		//Check if status hold this
		if _, ok := statusMap[key]; !ok {
			if patch == nil {
				patch = client.MergeFrom(c.object.DeepCopyObject().(client.Object))
			}
			status.Resources = append(status.Resources, &types.ResourceStatus{
				Name:  res.Name(),
				Type:  res.Type(),
				State: types.Creating,
			})
		}
	}
	if patch != nil {
		return ctrl.Result{Requeue: true}, c.cli.Status().Patch(ctx, c.object, patch)
	}

	for idx, res := range status.Resources {
		copyInstance := c.object.DeepCopyObject().(client.Object)
		if res == nil {
			continue
		}
		key := strings.ToLower(res.Type + "/" + res.Name)
		r, ok := resourceMap[key]
		if !ok {
			//TODO handle the case where we have a status for no resource. This can happen due to resource being removed from the list.
			continue
		}
		if err := c.reconcileObject(ctx, r, c.object, c.object.GetNamespace()); err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
		if converged, ok := c.object.(types.Converged); ok {
			if res.State != types.Waiting {
				status.Resources[idx].State = types.Waiting
				if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
					log.Error().Caller().Err(err).Msg("unable to update formation status")
					return ctrl.Result{}, err
				}
			}
			ok, err := converged.Converged(ctx, c.cli, c.object.GetNamespace())
			if err != nil {
				log.Error().Caller().Err(err).Msg("unable to check if formation is converged")
				return ctrl.Result{}, err
			}
			if !ok {
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
		}
		//Update the status
		status.Resources[idx].State = types.Ready
		if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
			log.Error().Caller().Err(err).Msg("unable to update formation status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// ReconcileObject In some cases, we want to reconcile a single object without the need to reconcile the whole formation.
// This will bypass the status check and will only check if the object exists and if not, create it.
// All the other logic will be the same as the Reconcile method.
func (c Controller) ReconcileObject(ctx context.Context, resource types.Resource, owner v1.Object) error {
	return c.reconcileObject(ctx, resource, owner, owner.GetNamespace())
}

func (c Controller) reconcileObject(ctx context.Context, resource types.Resource, owner v1.Object, namespace string) error {
	// Check if the resource is type.Reconcile, with some object like Secret, it might auto generate a new value on every reconcile.
	// To avoid this, the resource need to implement the type.Reconcile interface to handle its own reconcile.
	if r, ok := resource.(types.Reconcile); ok {
		return r.Reconcile(ctx, c.cli, namespace)
	}
	// get the resource from the API server
	instance := resource.Runtime()

	obj, err := resource.Create()
	obj.SetNamespace(namespace)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}
	if err := controllerutil.SetOwnerReference(owner, obj, c.scheme); err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}

	hash := HashObject(obj)
	if err := c.cli.Get(ctx, client.ObjectKey{Name: resource.Name(), Namespace: namespace}, instance); err != nil {
		if errors.IsNotFound(err) {
			//Create the instance
			annon := obj.GetAnnotations()
			if annon == nil {
				obj.SetAnnotations(map[string]string{})
				annon = obj.GetAnnotations()
			}
			annon[types.HashKey] = hash
			return c.cli.Create(ctx, obj)
		}
		log.Error().Caller().Err(err).Send()
		return err
	}
	instanceCopy := instance.DeepCopyObject()

	// Check if the hash match, this is done to reduce the amount of work we need to do going forward.
	annotations := instance.GetAnnotations()
	if h, ok := annotations[types.HashKey]; ok && h == hash {
		//Nothing changes
		return nil
	}

	if val, ok := annotations[types.UpdateKey]; ok && strings.ToLower(val) == "disabled" {
		return nil
	}

	annObj := obj.GetAnnotations()
	if annObj == nil {
		obj.SetAnnotations(map[string]string{})
		annObj = obj.GetAnnotations()
	}

	//If resource have custom Update logic, let that logic update the resource and create a patch from that.
	if sync, ok := resource.(types.Update); ok {
		obj2 := instance.DeepCopyObject()
		if err := sync.Update(ctx, obj2); err != nil {
			return err
		}
		obj = obj2.(client.Object)
	} else {
		if err = mergo.Merge(obj, instance); err != nil {
			return err
		}
	}
	obj.GetAnnotations()[types.HashKey] = hash

	jsonPatch, _ := jsonpatch.CreateJSONPatch(obj, instanceCopy)
	if len(jsonPatch.Raw()) < 2 {
		return nil
	}
	return c.cli.Patch(ctx, instance, client.RawPatch(k8sTypes.JSONPatchType, jsonPatch.Raw()))
}

// status.formation
func (c Controller) GetStatus() (*types.FormationStatus, error) {
	// Check if c.object is type FormationStatusInterface
	if status, ok := c.object.(types.FormationStatusInterface); ok {
		return status.GetStatus(), nil
	}
	//Use reflection to get the status from the object, this will assume the resource has a status field with a FormationStatus type.
	value, err := utils.GetValue2(c.object, "Status.Formation")
	if err != nil {
		return nil, errors2.New("unable to find formation status")
	}
	//The unsafe address is used to get the address of the value, this is needed to convert the value to a pointer.

	ptrToY := unsafe.Pointer(value.UnsafeAddr())
	return (*types.FormationStatus)(ptrToY), nil
}
