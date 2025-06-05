package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	// fruitMap := map[string]int{
	// 	"apple":  5,
	// 	"banana": 8,
	// }

	// // 检查存在的键
	// if count, exists := fruitMap["apple"]; exists {
	// 	fmt.Printf("苹果库存: %d\n", count) // 输出: 苹果库存: 5
	// } else {
	// 	fmt.Println("苹果缺货")
	// }

	// // 检查不存在的键
	// if count, exists := fruitMap["orange"]; exists {
	// 	fmt.Printf("橙子库存: %d\n", count)
	// } else {
	// 	fmt.Println("橙子缺货") // 输出: 橙子缺货
	// }

	// name := "linux-arm64-npu-1-7cksd-runner-7cf2c"
	// name = nil
	if count, err := extractNpuCountFromPodName(""); err != nil {
		fmt.Println(err)
	} else {
		fmt.Print(count)
	}
	fmt.Print("a")

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
