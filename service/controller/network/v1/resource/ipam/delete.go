package ipam

import (
	"context"
	"net"

	"github.com/giantswarm/cluster-operator/service/controller/network/v1/key"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
)

// EnsureDeleted takes care of freeing cluster subnet when ClusterNetworkConfig
// object is deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	clusterNetworkCfg, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(clusterNetworkCfg)

	if clusterNetworkCfg.Status.IP == "" {
		// Subnet not allocated. No need to do anything.
		r.logger.LogCtx(ctx, "level", "debug", "message", "Subnet not allocated. No need to do anything.", "clusterID", clusterID)
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Subnet allocated. Freeing.", "clusterID", clusterID)

	subnet := net.IPNet{
		IP:   net.ParseIP(clusterNetworkCfg.Status.IP),
		Mask: net.IPMask(net.ParseIP(clusterNetworkCfg.Status.Mask).To4()),
	}

	err = r.ipam.DeleteSubnet(ctx, subnet)
	if err != nil {
		return microerror.Maskf(err, "Subnet(%s, %s) freeing failed. clusterID: %s", subnet.IP.String(), net.IP(subnet.Mask).String(), clusterID)
	}

	clusterNetworkCfg.Status.IP = ""
	clusterNetworkCfg.Status.Mask = ""

	_, err = r.g8sClient.CoreV1alpha1().ClusterNetworkConfigs(accessor.GetNamespace()).UpdateStatus(&clusterNetworkCfg)
	if err != nil {
		return microerror.Maskf(err, "ClusterNetworkConfig status update failed. Subnet(%s, %s) is not allocated anymore but possibly used. clusterID: %s", subnet.IP.String(), subnet.Mask.String(), clusterID)
	}

	return nil
}
