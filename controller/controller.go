package controller

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"github.com/Doout/formation/internal/utils"
	"github.com/Doout/formation/types"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	jsonpatchv2 "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type Controller struct {
	cli    client.Client
	scheme *runtime.Scheme
	object client.Object

	transformers *Transformers
}

var rejectedPatchList = []string{
	//Reject all status patch change
	"/status",
	//Reject all metadata patch change that managed by Kubernetes
	"/metadata/creationTimestamp",
	"/metadata/generation",
	"/metadata/managedFields",
	"/metadata/uid",
	"/metadata/resourceVersion",
	"/metadata/selfLink",
}

func NewController(scheme *runtime.Scheme, cli client.Client) *Controller {
	return &Controller{scheme: scheme, cli: cli, transformers: NewTransformersWithDefault()}
}

func (c *Controller) AddTransformer(transformer ...mergo.Transformers) {
	c.transformers.Add(transformer...)
}

func (c Controller) ForObject(object client.Object) *Controller {
	b := &Controller{cli: c.cli, scheme: c.scheme, object: object, transformers: c.transformers}
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
	resourcesStatus := make([]*types.ResourceStatus, 0, len(list))
	for idx, res := range list {
		key := strings.ToLower(res.Type() + "/" + res.Name())
		resourceMap[key] = list[idx]
		//Check if status hold this
		if val, ok := statusMap[key]; ok {
			resourcesStatus = append(resourcesStatus, val)
			delete(statusMap, key)
		} else {
			if patch == nil {
				patch = client.MergeFrom(c.object.DeepCopyObject().(client.Object))
			}
			kinds, _, err := (*c.scheme).ObjectKinds(res.Runtime())
			if err != nil || len(kinds) == 0 {
				log.Error().Caller().Err(err).Msg("unable to get object kind")
				return ctrl.Result{}, err
			}
			rs := &types.ResourceStatus{
				Name:  res.Name(),
				Type:  res.Type(),
				Group: kinds[0].Group,
				State: types.Creating,
			}
			resourcesStatus = append(resourcesStatus, rs)
		}
	}
	//Add all the resources that are not in the list
	if len(statusMap) > 0 {
		leftOver := make([]*types.ResourceStatus, 0, len(statusMap))
		for _, val := range statusMap {
			leftOver = append(leftOver, val)
		}
		//Sort the left over, Map are not order and if we have any change in order; the status will be updated
		sort.Slice(leftOver, func(i, j int) bool {
			return leftOver[i].Name+leftOver[i].Type < leftOver[j].Name+leftOver[j].Type
		})
		resourcesStatus = append(resourcesStatus, leftOver...)
	}

	// Compare the status and the list, if there is a change, we need to update the status
	requestUpdate := len(status.Resources) != len(resourcesStatus)
	if !requestUpdate {
		// Check if all the resources are in the same order
		for idx, res := range status.Resources {
			if res == nil {
				continue
			}
			if res.Name != resourcesStatus[idx].Name || res.Type != resourcesStatus[idx].Type {
				requestUpdate = true
				break
			}
		}
	}

	if requestUpdate {
		patch = client.MergeFrom(c.object.DeepCopyObject().(client.Object))
	}

	if patch != nil {
		status.Resources = resourcesStatus
		return ctrl.Result{Requeue: true}, c.cli.Status().Patch(ctx, c.object, patch)
	}

	for idx := 0; idx < len(status.Resources); idx++ {
		res := status.Resources[idx]

		copyInstance := c.object.DeepCopyObject().(client.Object)
		if res == nil {
			continue
		}
		key := strings.ToLower(res.Type + "/" + res.Name)
		resource, ok := resourceMap[key]

		// Handle status found but no resource found, can happen due to the resource being removed from the resource list
		if !ok {
			// Get the unstructured resources matching the resource from status and delete them from the API server and status
			// If no resources are found, remove the resource from the status and continue
			resources := c.getUnstructuredObjects(res)

			// For each resource, delete it from the API server
			// Multiple resources can be returned if the resource has multiple versions
			for _, res := range resources {
				err := c.cli.Delete(ctx, res)
				// The resource could still be in use, or we don't have permission to delete it
				if err != nil && !errors.IsNotFound(err) {
					log.Error().Caller().Err(err).Msg("unable to delete resource")
					continue
				}
			}
			// Remove the resource from status if successfully deleted or not found
			removeResourceFromStatus(status, idx)
			if err = c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
				log.Error().Err(err).Msg("unable to update formation status")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, err
			}
			// Decrement the index to avoid skipping the next resource
			idx--
			continue
		}
		change, err := c.reconcileObject(ctx, resource, c.object, c.object.GetNamespace())
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}

		//Check if this resource is a part of ConvergedGroupInterface
		currentGroup := 0
		var convergedGroup types.ConvergedGroupInterface
		if convergedGroup, ok = resource.(types.ConvergedGroupInterface); ok {
			currentGroup = convergedGroup.GetConvergedGroupID()
		}
		nextGroupID := nextGroupIDFromList(status.Resources, resourceMap, idx)

		//If the object is not changed, we can skip the rest of the process
		if !change && res.State == types.Ready {
			if currentGroup == nextGroupID {
				continue
			}
			//Check if all previous are in ready state.
			if allPreviousStateReady(status, idx) {
				continue
			}
		}

		//Change there is some change, we need to update the status of this resource
		if res.State != types.Waiting {
			status.Resources[idx].State = types.Waiting
			if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
				log.Error().Caller().Err(err).Msg("unable to update formation status")
				return ctrl.Result{}, err
			}
		}

		if converged, ok := resource.(types.Converged); ok {
			if res.State != types.Waiting && res.State != types.Ready {
				status.Resources[idx].State = types.Waiting
				if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
					log.Error().Caller().Err(err).Msg("unable to update formation status")
					return ctrl.Result{}, err
				}
			}
			updateStatus, err := converged.Converged(ctx, c.cli, c.object.GetNamespace())
			if err != nil {
				log.Error().Caller().Err(err).Msg("unable to check if formation is converged")
				return ctrl.Result{}, err
			}
			// Update the status if needed, the rest of the logic below handle the group case
			if updateStatus {
				status.Resources[idx].State = types.Ready
				if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
					log.Error().Caller().Err(err).Msg("unable to update formation status")
					return ctrl.Result{}, err
				}
				//If the current group is 0, and this resource is converged, we can skip the rest of the logic
				if currentGroup == 0 {
					continue
				}
			}

			// Group id 0 mean no group and -1 mean no more resource. In both case we need to wait for the current resource to be converged
			if currentGroup == 0 || nextGroupID == -1 {
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
		} else {
			status.Resources[idx].State = types.Ready
			if err := c.cli.Status().Patch(ctx, c.object, client.MergeFrom(copyInstance)); err != nil {
				log.Error().Caller().Err(err).Msg("unable to update formation status")
				return ctrl.Result{}, err
			}
		}

		if nextGroupID != currentGroup {
			if !allPreviousStateReady(status, idx) {
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
		}
	}
	//Check if all resources are Ready
	for _, res := range status.Resources {
		if res == nil {
			continue
		}
		if res.State != types.Ready {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}
	return ctrl.Result{}, nil
}

// Check if all previous resources are ready
func allPreviousStateReady(status *types.FormationStatus, idx int) bool {
	// Check if all resources upto this point are converged,
	for i := 0; i <= idx; i++ {
		if status.Resources[i].State != types.Ready {
			return false
		}
	}
	return true
}

func nextGroupIDFromList(statusList []*types.ResourceStatus, resourceMap map[string]types.Resource, idx int) int {
	if idx+1 >= len(statusList) {
		return -1
	}
	nextResourceStatus := statusList[idx+1]
	nextKey := strings.ToLower(nextResourceStatus.Type + "/" + nextResourceStatus.Name)
	nextResource, ok := resourceMap[nextKey]
	if !ok {
		return 0
	}
	if convergedGroup, ok := nextResource.(types.ConvergedGroupInterface); ok {
		return convergedGroup.GetConvergedGroupID()
	}
	return 0
}

func removeResourceFromStatus(status *types.FormationStatus, idx int) {
	res := status.Resources[:idx]
	if idx+1 < len(status.Resources) {
		res = append(res, status.Resources[idx+1:]...)
	}
	status.Resources = res
}

// ReconcileObject In some cases, we want to reconcile a single object without the need to reconcile the whole formation.
// This will bypass the status check and will only check if the object exists and if not, create it.
// All the other logic will be the same as the Reconcile method.
func (c Controller) ReconcileObject(ctx context.Context, resource types.Resource, owner v1.Object) (bool, error) {
	return c.reconcileObject(ctx, resource, owner, owner.GetNamespace())
}

func (c Controller) createRuntimeObject(ctx context.Context, resource types.Resource, owner v1.Object, namespace string) (client.Object, error) {
	obj, err := resource.Create()
	obj.SetNamespace(namespace)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return nil, err
	}
	if err := controllerutil.SetOwnerReference(owner, obj, c.scheme); err != nil {
		log.Error().Caller().Err(err).Send()
		return nil, err
	}
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{})
	}
	return obj, nil
}

// getUnstructuredObjects returns a list of unstructured objects for the given resource status
func (c Controller) getUnstructuredObjects(res *types.ResourceStatus) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured

	for t := range c.scheme.AllKnownTypes() {
		if t.Kind == res.Type && t.Group == res.Group {
			resources = append(resources, c.getUnstructuredObject(res, &t))
			break
		}
	}
	return resources
}

// getUnstructuredObject returns an unstructured object for the given resource status and group version kind
func (c Controller) getUnstructuredObject(res *types.ResourceStatus, gvk *schema.GroupVersionKind) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvk.GroupVersion().String(),
			"kind":       gvk.Kind,
			"metadata": map[string]interface{}{
				"name":      res.Name,
				"namespace": c.object.GetNamespace(),
			},
		},
	}
}

func (c Controller) reconcileObject(ctx context.Context, resource types.Resource, owner v1.Object, namespace string) (bool, error) {
	// Check if the resource is type.Reconcile, with some object like Secret, it might auto generate a new value on every reconcile.
	// To avoid this, the resource need to implement the type.Reconcile interface to handle its own reconcile.
	if r, ok := resource.(types.Reconcile); ok {
		return r.Reconcile(ctx, c.cli, owner)
	}
	// get the resource from the API server
	instance := resource.Runtime()

	if err := c.cli.Get(ctx, client.ObjectKey{Name: resource.Name(), Namespace: namespace}, instance); err != nil {
		if errors.IsNotFound(err) {
			obj, err := c.createRuntimeObject(ctx, resource, owner, namespace)
			if err != nil {
				return false, err
			}
			obj.GetAnnotations()[types.HashKey] = HashObject(obj)
			return true, c.cli.Create(ctx, obj)
		}
		log.Error().Caller().Err(err).Send()
		return false, err
	}

	// Check if the hash match, this is done to reduce the amount of work we need to do going forward.
	annotations := instance.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if val, ok := annotations[types.UpdateKey]; ok && strings.ToLower(val) == "disabled" {
		return false, nil
	}
	// Create the runtime object
	obj, err := c.createRuntimeObject(ctx, resource, owner, namespace)
	if err != nil {
		return false, err
	}
	hash := HashObject(obj)
	if h, ok := annotations[types.HashKey]; ok && h == hash {
		//Nothing changes
		return false, nil
	}

	instanceCopy := instance.DeepCopyObject()
	//If resource have custom Update logic, let that logic update the resource and create a patch from that.
	if sync, ok := resource.(types.Update); ok {
		if err := sync.Update(ctx, instanceCopy); err != nil {
			return false, err
		}
		obj = instanceCopy.(client.Object)
	} else {
		if err = mergo.Merge(instanceCopy, obj, mergo.WithOverride, mergo.WithTransformers(c.transformers)); err != nil {
			return false, err
		}
		obj = instanceCopy.(client.Object)
	}

	obj.GetAnnotations()[types.HashKey] = hash
	objbytes, _ := json.Marshal(obj)
	instanceCopybytes, _ := json.Marshal(instance)

	jsonPatchs, _ := jsonpatchv2.CreatePatch(instanceCopybytes, objbytes)
	//We need at least 2 patch to be able to update the resource.
	// The first patch will be to update the hash annotations, the other patch will be to update the rest of the resource.
	totalNumberOfVaildPatch := 0
	rawPatch := "["

patch:
	for _, p := range jsonPatchs {
		for _, rej := range rejectedPatchList {
			//If the patch is in the rejected list, we skip it.
			if strings.HasPrefix(p.Path, rej) {
				continue patch
			}
		}
		totalNumberOfVaildPatch++
		rawPatch += p.Json() + ","
	}
	rawPatch = strings.TrimSuffix(rawPatch, ",")
	rawPatch += "]"
	// In some case; the hash will be the only thing that change, in this case, we don't want to update the resource.
	if totalNumberOfVaildPatch < 2 && strings.Contains(rawPatch, "replace") {
		return false, nil
	}
	if instance.GetObjectKind().GroupVersionKind().Kind != "Secret" {
		log.Debug().Caller().Str("instance", instance.GetName()).RawJSON("patch", []byte(rawPatch)).Msg("patch")
	} else {
		log.Debug().Caller().Str("instance", instance.GetName()).Msg("patch")
	}
	return true, c.cli.Patch(ctx, instance, client.RawPatch(k8sTypes.JSONPatchType, []byte(rawPatch)))
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
