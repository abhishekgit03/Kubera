package collector

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Collector struct {
	k8sClient          *kubernetes.Clientset
	metricsClient      *metricsv1beta1.Clientset
	mu                 sync.RWMutex
	latest             *ClusterSnapshot
	cpuCostPerCoreHour float64
	memCostPerGBHour   float64
	refreshInterval    time.Duration
}

func New(cpuCost, memCost float64, refreshInterval time.Duration) (*Collector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("k8s client: %w", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("metrics client: %w", err)
	}

	c := &Collector{
		k8sClient:          k8sClient,
		metricsClient:      metricsClient,
		cpuCostPerCoreHour: cpuCost,
		memCostPerGBHour:   memCost,
		refreshInterval:    refreshInterval,
	}

	// Initial collection
	c.collect(context.Background())

	// Background ticker
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			c.collect(context.Background())
		}
	}()

	return c, nil
}

func (c *Collector) Latest() *ClusterSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}

func (c *Collector) collect(ctx context.Context) {
	snap, err := c.doCollect(ctx)
	if err != nil {
		log.Printf("collection error: %v", err)
		return
	}
	c.mu.Lock()
	c.latest = snap
	c.mu.Unlock()
	log.Printf("collected snapshot: %d workloads, metrics_available=%v", len(snap.Workloads), snap.MetricsAvailable)
}

func (c *Collector) doCollect(ctx context.Context) (*ClusterSnapshot, error) {
	nsList, err := c.k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	snap := &ClusterSnapshot{
		CollectedAt:     time.Now().UTC().Format(time.RFC3339),
		TotalNamespaces: len(nsList.Items),
		MetricsAvailable: true,
	}

	// Map workload uid -> WorkloadCost for aggregation
	workloads := map[string]*WorkloadCost{}
	// Map pod name -> workload uid for metric lookup
	podToWorkload := map[string]string{}

	totalPods := 0

	for _, ns := range nsList.Items {
		nsName := ns.Name

		// Collect Deployments
		deps, err := c.k8sClient.AppsV1().Deployments(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("list deployments in %s: %v", nsName, err)
		} else {
			for i := range deps.Items {
				d := &deps.Items[i]
				uid := string(d.UID)
				replicas := int32(1)
				if d.Spec.Replicas != nil {
					replicas = *d.Spec.Replicas
				}
				wc := &WorkloadCost{
					Namespace:    nsName,
					WorkloadName: d.Name,
					WorkloadType: "Deployment",
					Replicas:     replicas,
				}
				accumulateContainerResources(wc, d.Spec.Template.Spec.Containers)
				workloads[uid] = wc
			}
		}

		// Collect StatefulSets
		ssets, err := c.k8sClient.AppsV1().StatefulSets(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("list statefulsets in %s: %v", nsName, err)
		} else {
			for i := range ssets.Items {
				s := &ssets.Items[i]
				uid := string(s.UID)
				replicas := int32(1)
				if s.Spec.Replicas != nil {
					replicas = *s.Spec.Replicas
				}
				wc := &WorkloadCost{
					Namespace:    nsName,
					WorkloadName: s.Name,
					WorkloadType: "StatefulSet",
					Replicas:     replicas,
				}
				accumulateContainerResources(wc, s.Spec.Template.Spec.Containers)
				workloads[uid] = wc
			}
		}

		// Collect DaemonSets
		dsets, err := c.k8sClient.AppsV1().DaemonSets(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("list daemonsets in %s: %v", nsName, err)
		} else {
			for i := range dsets.Items {
				d := &dsets.Items[i]
				uid := string(d.UID)
				wc := &WorkloadCost{
					Namespace:    nsName,
					WorkloadName: d.Name,
					WorkloadType: "DaemonSet",
					Replicas:     d.Status.DesiredNumberScheduled,
				}
				accumulateContainerResources(wc, d.Spec.Template.Spec.Containers)
				workloads[uid] = wc
			}
		}

		// Map pods to workloads (two levels: pod -> RS -> Deployment)
		pods, err := c.k8sClient.CoreV1().Pods(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("list pods in %s: %v", nsName, err)
			continue
		}
		totalPods += len(pods.Items)

		// Build RS -> Deployment map
		rsToDeployment := map[string]string{}
		rsets, err := c.k8sClient.AppsV1().ReplicaSets(nsName).List(ctx, metav1.ListOptions{})
		if err == nil {
			for i := range rsets.Items {
				rs := &rsets.Items[i]
				for _, owner := range rs.OwnerReferences {
					if owner.Kind == "Deployment" {
						rsToDeployment[string(rs.UID)] = string(owner.UID)
					}
				}
			}
		}

		for i := range pods.Items {
			pod := &pods.Items[i]
			for _, owner := range pod.OwnerReferences {
				switch owner.Kind {
				case "ReplicaSet":
					if depUID, ok := rsToDeployment[string(owner.UID)]; ok {
						podToWorkload[pod.Name+"/"+nsName] = depUID
					} else {
						podToWorkload[pod.Name+"/"+nsName] = string(owner.UID)
					}
				case "StatefulSet", "DaemonSet":
					podToWorkload[pod.Name+"/"+nsName] = string(owner.UID)
				}
			}
		}
	}

	// Fetch pod metrics and aggregate to workloads
	for _, ns := range nsList.Items {
		podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			snap.MetricsAvailable = false
			log.Printf("metrics unavailable for %s: %v", ns.Name, err)
			break
		}

		for i := range podMetrics.Items {
			pm := &podMetrics.Items[i]
			key := pm.Name + "/" + ns.Name
			workloadUID, ok := podToWorkload[key]
			if !ok {
				continue
			}
			wc, ok := workloads[workloadUID]
			if !ok {
				continue
			}
			for _, cm := range pm.Containers {
				wc.CPUActualMillis += cm.Usage.Cpu().MilliValue()
				wc.MemActualMB += cm.Usage.Memory().Value() / (1024 * 1024)
			}
		}
	}

	// Flatten workloads, compute cost/waste fields, accumulate totals
	const hoursPerMonth = 730.0
	for _, wc := range workloads {
		// CPU waste ratio: actual/request (1.0 means fully utilized, low = wasted)
		if wc.CPURequestMillis > 0 {
			r := float64(wc.CPUActualMillis) / float64(wc.CPURequestMillis)
			if r > 1.0 {
				r = 1.0
			}
			wc.CPUWasteRatio = r
		} else {
			wc.CPUWasteRatio = 1.0
		}
		if wc.MemRequestMB > 0 {
			r := float64(wc.MemActualMB) / float64(wc.MemRequestMB)
			if r > 1.0 {
				r = 1.0
			}
			wc.MemWasteRatio = r
		} else {
			wc.MemWasteRatio = 1.0
		}

		cpuCores := float64(wc.CPURequestMillis) / 1000.0
		memGB := float64(wc.MemRequestMB) / 1024.0
		rep := float64(wc.Replicas)
		wc.EstimatedMonthlyCost = (cpuCores*c.cpuCostPerCoreHour + memGB*c.memCostPerGBHour) * rep * hoursPerMonth

		idleCPU := float64(wc.CPURequestMillis-wc.CPUActualMillis) / 1000.0
		idleMem := float64(wc.MemRequestMB-wc.MemActualMB) / 1024.0
		if idleCPU < 0 {
			idleCPU = 0
		}
		if idleMem < 0 {
			idleMem = 0
		}
		wc.EstimatedMonthlyWaste = (idleCPU*c.cpuCostPerCoreHour + idleMem*c.memCostPerGBHour) * rep * hoursPerMonth

		snap.Workloads = append(snap.Workloads, *wc)
		snap.TotalEstimatedCost += wc.EstimatedMonthlyCost
		snap.TotalEstimatedWaste += wc.EstimatedMonthlyWaste
	}
	snap.TotalWorkloads = len(snap.Workloads)
	snap.TotalPods = totalPods

	return snap, nil
}

func accumulateContainerResources(wc *WorkloadCost, containers []corev1.Container) {
	for _, c := range containers {
		if req := c.Resources.Requests; req != nil {
			if cpu := req.Cpu(); cpu != nil {
				wc.CPURequestMillis += cpu.MilliValue()
			}
			if mem := req.Memory(); mem != nil {
				wc.MemRequestMB += mem.Value() / (1024 * 1024)
			}
		}
		if lim := c.Resources.Limits; lim != nil {
			if cpu := lim.Cpu(); cpu != nil {
				wc.CPULimitMillis += cpu.MilliValue()
			}
			if mem := lim.Memory(); mem != nil {
				wc.MemLimitMB += mem.Value() / (1024 * 1024)
			}
		}
	}
}
