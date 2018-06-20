package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/service/controller/network/v1/key"
)

// EnsureCreated takes care of cluster subnet allocation when
// ClusterNetworkConfig object is created.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	clusterNetworkCfg, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(clusterNetworkCfg)

	r.logger.LogCtx(ctx, "level", "debug", "message", "allocating subnet", "clusterID", clusterID)

	if clusterNetworkCfg.Status.IP != "" {
		// Subnet allocated. No need to do anything.
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to allocate subnet", "clusterID", clusterID)
		return nil
	}

	maskBits := key.ClusterNetworkMaskBits(clusterNetworkCfg)

	mask := net.CIDRMask(maskBits, 32)

	subnet, err := r.ipam.CreateSubnet(ctx, mask, clusterID)
	if err != nil {
		return microerror.Maskf(err, "subnet allocation failed with mask %s, clusterID: %s", net.IP(mask).String(), clusterID)
	}

	clusterNetworkCfg.Status.IP = subnet.IP.String()
	clusterNetworkCfg.Status.Mask = net.IP(subnet.Mask).String()

	_, err = r.g8sClient.CoreV1alpha1().ClusterNetworkConfigs(clusterNetworkCfg.GetNamespace()).UpdateStatus(&clusterNetworkCfg)
	if err != nil {
		ipamDeleteErr := r.ipam.DeleteSubnet(ctx, subnet)
		if ipamDeleteErr != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("freeing subnet(%s, %s) failed", subnet.IP.String(), subnet.Mask.String()), "clusterID", clusterID, "stack", fmt.Sprintf("%#v", ipamDeleteErr))
			return microerror.Maskf(err, "ClusterNetworkConfig status update failed for clusterID %s. Freeing Subnet(%s, %s) allocation also failed. It is allocated but possibly not used.", clusterID, subnet.IP.String(), subnet.Mask.String())
		}
		return microerror.Maskf(err, "ClusterNetworkConfig status update failed for clusterID %s. Allocated Subnet(%s, %s) freed.", clusterID, subnet.IP.String(), subnet.Mask.String())
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocated subnet(%s, %s)", clusterNetworkCfg.Status.IP, clusterNetworkCfg.Status.Mask), "clusterID", clusterID)

	return nil
}
