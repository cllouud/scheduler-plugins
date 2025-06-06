package runnernpu

import (
	"context"
	"fmt"
	"log"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name          = "labelAB"
	npuCountLabel = "ascend-ci.com/required-npu-count"
)

type labelAB struct{}

var _ framework.FilterPlugin = &labelAB{}
var _ framework.PreScorePlugin = &labelAB{}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &labelAB{}, nil
}

// Name returns name of the plugin.
func (pl *labelAB) Name() string {
	return Name
}

func (pl *labelAB) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	schedulingPodNpuCount, exists, err := extractNpuCountFromPodLabel(pod)
	if !exists {
		return framework.NewStatus(framework.Success, "")
	}
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, err.Error())
	}

	allocatedNpuCount := 0
	for _, podInfo := range nodeInfo.Pods {
		podNpuCount, exists, err := extractNpuCountFromPodLabel(podInfo.Pod)
		if !exists {
			continue
		}
		if err != nil {
			return framework.NewStatus(framework.Unschedulable, err.Error())
		}
		allocatedNpuCount += podNpuCount

		log.Printf("node: %v, pod: %v, podNpuCount: %v, allocatedNpuCount: %v, schePod: %v, scheCount: %v", nodeInfo.Node().Name, podInfo.Pod.Name, podNpuCount, allocatedNpuCount, pod.Name, schedulingPodNpuCount)
	}

	allocatableNpuCount, ok := nodeInfo.Allocatable.ScalarResources["huawei.com/ascend-1980"]
	if !ok {
		return framework.NewStatus(framework.Unschedulable, "can not get allocatable_npu_count from node")
	}

	names := make([]string, 0, len(nodeInfo.Pods))
	for _, pod := range nodeInfo.Pods {
		if pod.Pod != nil {
			names = append(names, pod.Pod.Name)
		}
	}

	log.Printf("node: %v, allocatedNpuCount: %v, schePod: %v, scheCount: %v, pod names: %v", nodeInfo.Node().Name, allocatedNpuCount, pod.Name, schedulingPodNpuCount, names)
	if allocatableNpuCount-int64(allocatedNpuCount) < int64(schedulingPodNpuCount) {
		return framework.NewStatus(framework.Unschedulable, "current node has no enough npu")
	}

	return framework.NewStatus(framework.Success, "")

}

func (pl *labelAB) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	log.Printf("PreScore: %+v", nodes)
	return framework.NewStatus(framework.Success, "Pod: "+pod.Name)
}

func extractNpuCountFromPodLabel(pod *v1.Pod) (int, bool, error) {
	labelValue, exists := pod.Labels[npuCountLabel]
	if !exists {
		return 0, exists, nil
	}

	npuCount, err := strconv.Atoi(labelValue)
	if err != nil {
		return 0, exists, fmt.Errorf("failed to parse NPU count, pod: %v, label: %v", pod, pod.Labels[npuCountLabel])
	}
	return npuCount, exists, nil
}
