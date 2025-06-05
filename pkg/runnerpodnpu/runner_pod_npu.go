package runnernpu

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "labelAB"

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
	desire_npu_count, err := extractNpuCountFromPodName(pod.Name)
	if err != nil {
		log.Printf("extractNpuCountFromPodName error: %v", err)
	}

	log.Printf("node: %v, nodeInfo: %v", nodeInfo.Node().Name, nodeInfo)

	allocatable_npu_count, ok := nodeInfo.Allocatable.ScalarResources["huawei.com/ascend-1980"]
	if !ok {
		log.Printf("can not get allocatable_npu_count from node")
	}

	requested_npu_count, ok := nodeInfo.Requested.ScalarResources["huawei.com/ascend-1980"]
	if !ok {
		log.Printf("can not get requested_npu_count from node")
	}

	log.Printf("desire_npu_count: %v, pod_name: %v, allocatable_npu_count: %v, requested_npu_count: %v",
		desire_npu_count, pod.Name, allocatable_npu_count, requested_npu_count)

	if int64(desire_npu_count) <= allocatable_npu_count-requested_npu_count {
		return framework.NewStatus(framework.Success, "")
	} else {
		return framework.NewStatus(framework.Unschedulable, "npu count not enough")
	}
}

func (pl *labelAB) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	log.Printf("PreScore: %+v", nodes)
	return framework.NewStatus(framework.Success, "Pod: "+pod.Name)
}

func extractNpuCountFromPodName(name string) (int, error) {
	parts := strings.Split(name, "-")
	if len(parts) < 4 {
		return -1, fmt.Errorf("can not extract name: %s", name)
	}

	if count, err := strconv.Atoi(parts[3]); err != nil {
		return -1, fmt.Errorf("can not extract name: %s", name)
	} else {
		return count, nil
	}
}
