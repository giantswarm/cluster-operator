// NOTE this file is copied from operatorkit for migration purposes. The goal
// here is to get rid of the crud primitive and move to the implementation of
// the new handler interface eventually.
package certconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/loggermeta"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/resourcecanceledcontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error

	var currentState interface{}
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "GetCurrentState"
			defer delete(meta.KeyVals, "function")
		}
		currentState, err = r.getCurrentState(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var desiredState interface{}
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "GetDesiredState"
			defer delete(meta.KeyVals, "function")
		}
		desiredState, err = r.getDesiredState(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var patch *patch
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "NewUpdatePatch"
			defer delete(meta.KeyVals, "function")
		}
		patch, err = r.newUpdatePatch(ctx, obj, currentState, desiredState)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			createState, ok := patch.getCreateChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyCreateChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyCreateChange(ctx, obj, createState)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			deleteState, ok := patch.getDeleteChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyDeleteChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyDeleteChange(ctx, obj, deleteState)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			updateState, ok := patch.getUpdateChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyUpdateChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyUpdateChange(ctx, obj, updateState)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	var err error

	var currentState interface{}
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "GetCurrentState"
			defer delete(meta.KeyVals, "function")
		}
		currentState, err = r.getCurrentState(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var desiredState interface{}
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "GetDesiredState"
			defer delete(meta.KeyVals, "function")
		}
		desiredState, err = r.getDesiredState(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var patch *patch
	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		meta, ok := loggermeta.FromContext(ctx)
		if ok {
			meta.KeyVals["function"] = "NewDeletePatch"
			defer delete(meta.KeyVals, "function")
		}
		patch, err = r.newDeletePatch(ctx, obj, currentState, desiredState)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			createChange, ok := patch.getCreateChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyCreateChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyCreateChange(ctx, obj, createChange)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			deleteChange, ok := patch.getDeleteChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyDeleteChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyDeleteChange(ctx, obj, deleteChange)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	{
		if reconciliationcanceledcontext.IsCanceled(ctx) {
			return nil
		}
		if resourcecanceledcontext.IsCanceled(ctx) {
			return nil
		}

		if patch != nil {
			updateChange, ok := patch.getUpdateChange()
			if ok {
				meta, ok := loggermeta.FromContext(ctx)
				if ok {
					meta.KeyVals["function"] = "ApplyUpdateChange"
					defer delete(meta.KeyVals, "function")
				}
				err := r.applyUpdateChange(ctx, obj, updateChange)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	return nil
}
