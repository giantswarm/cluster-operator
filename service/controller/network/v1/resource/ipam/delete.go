package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/cluster-operator/service/controller/network/v1/key"
	"github.com/giantswarm/microerror"
)

// EnsureDeleted takes care of freeing cluster subnet when ClusterNetworkConfig
// object is deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	clusterNetworkCfg, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(clusterNetworkCfg)

	r.logger.LogCtx(ctx, "level", "debug", "message", "freeing subnet", "clusterID", clusterID)

	if clusterNetworkCfg.Status.IP == "" {
		// Subnet not allocated. No need to do anything.
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to free subnet", "clusterID", clusterID)
		return nil
	}

	subnet := net.IPNet{
		IP:   net.ParseIP(clusterNetworkCfg.Status.IP),
		Mask: net.IPMask(net.ParseIP(clusterNetworkCfg.Status.Mask).To4()),
	}

	err = r.ipam.DeleteSubnet(ctx, subnet)
	if err != nil {
		return microerror.Maskf(err, "Subnet(%s, %s) freeing failed, clusterID: %s", subnet.IP.String(), net.IP(subnet.Mask).String(), clusterID)
	}

	clusterNetworkCfg.Status.IP = ""
	clusterNetworkCfg.Status.Mask = ""

	_, err = r.g8sClient.CoreV1alpha1().ClusterNetworkConfigs(clusterNetworkCfg.GetNamespace()).UpdateStatus(&clusterNetworkCfg)
	if err != nil {
		return microerror.Maskf(err, "ClusterNetworkConfig status update failed. Subnet(%s, %s) is not allocated anymore but possibly used. clusterID: %s", subnet.IP.String(), subnet.Mask.String(), clusterID)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("freed subnet(%s, %s)", subnet.IP.String(), net.IP(subnet.Mask).String()), "clusterID", clusterID)

	return nil
}
