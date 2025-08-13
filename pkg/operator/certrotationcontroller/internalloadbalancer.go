package certrotationcontroller

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *CertRotationController) syncInternalLoadBalancerHostnames(ctx context.Context) error {
	infrastructureConfig, err := c.infrastructureLister.Get(ctx, "cluster")
	if err != nil {
		return err
	}
	hostname := infrastructureConfig.Status.APIServerInternalURL
	// if hostname is not set do not panic with slice bounds out of range
	if len(hostname) == 0 {
		klog.WarningfWithCtx(ctx, "Failed to set internal loadbalancer: APIServerInternalURL is not set")
		return nil
	}
	hostname = strings.Replace(hostname, "https://", "", 1)
	hostname = hostname[0:strings.LastIndex(hostname, ":")]

	klog.V(2).InfofWithCtx(ctx, "syncing internal loadbalancer hostnames: %v", hostname)
	c.internalLoadBalancer.setHostnames([]string{hostname})
	return nil
}

func (c *CertRotationController) runInternalLoadBalancerHostnames(ctx context.Context) {
	for c.processInternalLoadBalancerHostnames(ctx) {
	}
}

func (c *CertRotationController) processInternalLoadBalancerHostnames(ctx context.Context) bool {
	dsKey, quit := c.internalLoadBalancerHostnamesQueue.Get()
	if quit {
		return false
	}
	defer c.internalLoadBalancerHostnamesQueue.Done(dsKey)

	err := c.syncInternalLoadBalancerHostnames(ctx)
	if err == nil {
		c.internalLoadBalancerHostnamesQueue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.internalLoadBalancerHostnamesQueue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *CertRotationController) internalLoadBalancerHostnameEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.internalLoadBalancerHostnamesQueue.Add(workQueueKey) },
	}
}
