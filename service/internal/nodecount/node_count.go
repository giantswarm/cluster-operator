package nodecount

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/cluster-operator/pkg/label"
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
	o := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s", label.MasterNodeRole),
	}
	nodes, err := nc.cachedNodes(ctx, o, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Here we can filter for the master nodes.

	masterCount := make(map[string]Node)
	for _, node := range nodes.Items {
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

	return masterCount, nil
}

func (nc *NodeCount) WorkerCount(ctx context.Context, obj interface{}) (map[string]Node, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	o := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("!%s", label.MasterNodeRole),
	}

	nodes, err := nc.cachedNodes(ctx, o, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Here we can filter for the worker nodes.

	workerCount := make(map[string]Node)
	for _, node := range nodes.Items {
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

	return workerCount, nil
}

func (nc *NodeCount) cachedNodes(ctx context.Context, o metav1.ListOptions, cr metav1.Object) (corev1.NodeList, error) {
	var err error
	var ok bool

	var nodes corev1.NodeList
	{
		ck := nc.nodesCache.Key(ctx, cr)

		if ck == "" {
			nodes, err = nc.lookupNodes(o)
			if err != nil {
				return corev1.NodeList{}, microerror.Mask(err)
			}
		} else {
			nodes, ok = nc.nodesCache.Get(ctx, ck)
			if !ok {
				nodes, err = nc.lookupNodes(o)
				if err != nil {
					return corev1.NodeList{}, microerror.Mask(err)
				}

				nc.nodesCache.Set(ctx, ck, nodes)
			}
		}
	}

	return nodes, nil
}

func (nc *NodeCount) lookupNodes(cr runtime.Object) (corev1.NodeList, error) {
	// Here is where we need to create the Tenant Cluster Kubernetes client
	// using the CR.

	// TODO get tenant client
	//
	//     1. take it from controller context
	//     2. make it up ourselves the right way
	//     3. copy some code over here
	//

	// We should make up the list options in here.
	nodes, err := k8sClient.K8sClient().CoreV1().Nodes().List(o)
	if err != nil {
		return corev1.NodeList{}, microerror.Mask(err)
	}

	if len(nodes.Items) == 0 {
		return corev1.NodeList{}, microerror.Mask(notFoundError)
	}
	return *nodes, nil
}
