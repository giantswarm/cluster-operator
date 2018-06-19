package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/cluster-operator/service/controller/network/v1/key"
)

// EnsureCreated takes care of cluster subnet allocation when
// ClusterNetworkConfig object is created.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	clusterNetworkCfg, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(clusterNetworkCfg)

	if clusterNetworkCfg.Status.IP != "" {
		// Subnet allocated. No need to do anything.
		r.logger.LogCtx(ctx, "level", "debug", "message", "Subnet allocated. No need to do anything.", "clusterID", clusterID)
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Subnet not allocated. Allocating.", "clusterID", clusterID)

	maskBits := key.ClusterNetworkMaskBits(clusterNetworkCfg)

	mask := net.CIDRMask(maskBits, 32)

	subnet, err := r.ipam.CreateSubnet(ctx, mask, clusterID)
	if err != nil {
		return microerror.Maskf(err, "Subnet allocation failed with mask %s. clusterID: %s", net.IP(mask).String(), clusterID)
	}

	clusterNetworkCfg.Status.IP = subnet.IP.String()
	clusterNetworkCfg.Status.Mask = net.IP(subnet.Mask).String()

	_, err = r.g8sClient.CoreV1alpha1().ClusterNetworkConfigs(accessor.GetNamespace()).UpdateStatus(&clusterNetworkCfg)
	if err != nil {
		ipamDeleteErr := r.ipam.DeleteSubnet(ctx, subnet)
		if ipamDeleteErr != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("Freeing Subnet(%s, %s) failed.", subnet.IP.String(), subnet.Mask.String()), "clusterID", clusterID, "stack", fmt.Sprintf("%#v", ipamDeleteErr))
			return microerror.Maskf(err, "ClusterNetworkConfig status update failed for clusterID %s. Freeing Subnet(%s, %s) allocation also failed. It is allocated but possibly not used.", clusterID, subnet.IP.String(), subnet.Mask.String())
		}
		return microerror.Maskf(err, "ClusterNetworkConfig status update failed for clusterID %s. Allocated Subnet(%s, %s) successfully freed.", clusterID, subnet.IP.String(), subnet.Mask.String())
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Subnet allocated: %s, %s.", clusterNetworkCfg.Status.IP, clusterNetworkCfg.Status.Mask), "clusterID", clusterID)

	return nil
}
