package runnernpu

import (
	"context"
	"fmt"
	"log"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name          = "runnerScheduler"
	npuCountLabel = "ascend-ci.com/required-npu-count"
)

type runnerScheduler struct{}

var _ framework.FilterPlugin = &runnerScheduler{}
var _ framework.PreScorePlugin = &runnerScheduler{}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &runnerScheduler{}, nil
}

// Name returns name of the plugin.
func (pl *runnerScheduler) Name() string {
	return Name
}

// 根据待调度pod与当前节点的剩余NPU卡判断是否可以调度到当前节点。
// 调度pod的`label.ascend-ci.com/required-npu-count`表明pod所需NPU卡。
// 当前节点的`allocatable.huawei.com/ascend-1980`表明当前节点的总卡数。
// 遍历当前节点的所有pod，将其`label.ascend-ci.com/required-npu-count`相加，作为当前节点已分配卡数。
// 如果当前节点的总卡数-当前节点已分配卡数<=pod所需NPU卡，则可以将pod分配到当前节点。
func (pl *runnerScheduler) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	schedulingPodNpuCount, exists, err := extractNpuCountFromPodLabel(pod)
	if !exists {
		return framework.NewStatus(framework.Success, "")
	}
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, err.Error())
	}

	allocatedNpuCount := 0
	podNames := make([]string, 0, len(nodeInfo.Pods))
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod != nil {
			podNames = append(podNames, podInfo.Pod.Name)
		}
		podNpuCount, exists, err := extractNpuCountFromPodLabel(podInfo.Pod)
		if !exists || err != nil {
			continue
		}
		allocatedNpuCount += podNpuCount

		klog.V(4).InfoS("node: %v, pod: %v, podNpuCount: %v, allocatedNpuCount: %v, schePod: %v, scheCount: %v", nodeInfo.Node().Name, podInfo.Pod.Name, podNpuCount, allocatedNpuCount, pod.Name, schedulingPodNpuCount)
	}

	allocatableNpuCount, ok := nodeInfo.Allocatable.ScalarResources["huawei.com/ascend-1980"]
	if !ok {
		return framework.NewStatus(framework.Unschedulable, "can not get allocatable_npu_count from node")
	}

	klog.InfoS("Node status", "nodeName", nodeInfo.Node().Name, "allocatableNpu", allocatableNpuCount, "allocatedNpu", allocatedNpuCount, "schedulingCount", schedulingPodNpuCount, "schedulingPod", pod.Name, "nodePodsNames", podNames)

	if allocatableNpuCount-int64(allocatedNpuCount) < int64(schedulingPodNpuCount) {
		klog.Infof("current node has no enough npu, node name : %v", nodeInfo.Node().Name)
		return framework.NewStatus(framework.Unschedulable, "current node has no enough npu")
	}

	return framework.NewStatus(framework.Success, "")

}

func (pl *runnerScheduler) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
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
