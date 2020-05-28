package nodecount

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/internal/nodecount/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type NodeCount struct {
	k8sClient k8sclient.Interface

	nodesCache *cache.Nodes
}

func New(c Config) (*NodeCount, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	nc := &NodeCount{
		k8sClient: c.K8sClient,

		nodesCache: cache.NewNodes(),
	}

	return nc, nil
}

func (nc *NodeCount) MasterCount(ctx context.Context, obj interface{}) (map[string]Node, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	nodes, err := nc.cachedNodes(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	masterCount := make(map[string]Node)
	for _, node := range nodes.Items {
		if _, ok := node.Labels[label.MasterNodeRole]; ok {
			id := node.Labels[label.ControlPlane]
			{
				val := masterCount[id]
				val.Nodes++
				masterCount[id] = val
			}
			for _, c := range node.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					val := masterCount[id]
					val.Ready++
					masterCount[id] = val
				}
			}
		}
	}

	return masterCount, nil
}

func (nc *NodeCount) WorkerCount(ctx context.Context, obj interface{}) (map[string]Node, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	nodes, err := nc.cachedNodes(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	workerCount := make(map[string]Node)
	for _, node := range nodes.Items {
		if _, ok := node.Labels[label.WorkerNodeRole]; ok {
			id := node.Labels[label.MachineDeployment]
			{
				val := workerCount[id]
				val.Nodes++
				workerCount[id] = val
			}
			for _, c := range node.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					val := workerCount[id]
					val.Ready++
					workerCount[id] = val
				}
			}
		}
	}

	return workerCount, nil
}

func (nc *NodeCount) cachedNodes(ctx context.Context, cr metav1.Object) (corev1.NodeList, error) {
	var err error
	var ok bool

	var nodes corev1.NodeList
	{
		ck := nc.nodesCache.Key(ctx, cr)

		if ck == "" {
			nodes, err = nc.lookupNodes(ctx)
			if err != nil {
				return corev1.NodeList{}, microerror.Mask(err)
			}
		} else {
			nodes, ok = nc.nodesCache.Get(ctx, ck)
			if !ok {
				nodes, err = nc.lookupNodes(ctx)
				if err != nil {
					return corev1.NodeList{}, microerror.Mask(err)
				}

				nc.nodesCache.Set(ctx, ck, nodes)
			}
		}
	}

	return nodes, nil
}

func (nc *NodeCount) lookupNodes(ctx context.Context) (corev1.NodeList, error) {
	// TODO we need to get rid off the controllercontext.Context but for now it should be fine.
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return corev1.NodeList{}, microerror.Mask(err)
	}
	if cc.Client.TenantCluster.K8s != nil {
		nodes, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return corev1.NodeList{}, microerror.Mask(err)
		}

		if len(nodes.Items) == 0 {
			return corev1.NodeList{}, microerror.Mask(notFoundError)
		}
		return *nodes, nil
	}
	return corev1.NodeList{}, microerror.Mask(interfaceError)
}
