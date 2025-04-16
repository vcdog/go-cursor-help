package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConfigStructure represents the structure of Cursor's config
type ConfigStructure struct {
	Telemetry struct {
		MachineMachineId string `json:"machineMachineId,omitempty"`
		MacMachineId     string `json:"macMachineId,omitempty"`
		WinMachineId     string `json:"winMachineId,omitempty"`
		LinuxMachineId   string `json:"linuxMachineId,omitempty"`
	} `json:"telemetry"`
}

func main() {
	// 打印当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前工作目录失败: %v\n", err)
		return
	}
	fmt.Printf("当前工作目录: %s\n", dir)

	// 可能的配置文件位置
	configLocations := []string{
		filepath.Join(dir, "settings.json"),
		filepath.Join(dir, "config.json"),
		filepath.Join(dir, ".cursor", "settings.json"),
		filepath.Join(dir, ".vscode", "settings.json"),
		filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "settings.json"),
		filepath.Join(os.Getenv("USERPROFILE"), ".cursor", "settings.json"),
	}

	// 搜索配置文件
	var configPath string
	for _, path := range configLocations {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			fmt.Printf("找到配置文件: %s\n", configPath)
			break
		}
	}

	if configPath == "" {
		fmt.Println("未找到配置文件")
		return
	}

	// 读取配置文件
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		return
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		return
	}

	// 检查配置结构
	fmt.Println("配置文件内容:")
	printMap(config, 0)

	// 修复 telemetry 配置
	fixed, newConfig := fixTelemetryConfig(config)
	if !fixed {
		fmt.Println("无需修复配置")
		return
	}

	// 保存修复后的配置
	newConfigData, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		fmt.Printf("序列化新配置失败: %v\n", err)
		return
	}

	// 备份原始配置
	backupPath := configPath + ".backup"
	if err := ioutil.WriteFile(backupPath, configData, 0644); err != nil {
		fmt.Printf("备份原始配置失败: %v\n", err)
		return
	}
	fmt.Printf("已备份原始配置到: %s\n", backupPath)

	// 写入新配置
	if err := ioutil.WriteFile(configPath, newConfigData, 0644); err != nil {
		fmt.Printf("写入新配置失败: %v\n", err)
		return
	}
	fmt.Printf("已成功修复并保存配置到: %s\n", configPath)
}

// fixTelemetryConfig 修复 telemetry 配置结构
func fixTelemetryConfig(config map[string]interface{}) (bool, map[string]interface{}) {
	// 深拷贝配置
	newConfig := make(map[string]interface{})
	for k, v := range config {
		newConfig[k] = v
	}

	// 检查 telemetry 对象是否存在
	telemetry, exists := newConfig["telemetry"].(map[string]interface{})
	if !exists {
		// 创建 telemetry 对象
		telemetry = make(map[string]interface{})
		newConfig["telemetry"] = telemetry
		fmt.Println("创建了 telemetry 对象")
	}

	// 获取当前 machineId
	var machineId string
	// 检查平台特定的 machineId
	switch runtime.GOOS {
	case "darwin":
		if id, ok := telemetry["macMachineId"].(string); ok && id != "" {
			machineId = id
		}
	case "windows":
		if id, ok := telemetry["winMachineId"].(string); ok && id != "" {
			machineId = id
		}
	case "linux":
		if id, ok := telemetry["linuxMachineId"].(string); ok && id != "" {
			machineId = id
		}
	}

	// 如果没有找到平台特定的 machineId，使用通用的 machineMachineId
	if machineId == "" {
		if id, ok := telemetry["machineMachineId"].(string); ok && id != "" {
			machineId = id
		}
	}

	// 如果仍然没有 machineId，生成一个新的
	if machineId == "" {
		machineId = generateMachineId()
	}

	// 确保所有平台的 machineId 都设置了
	telemetryMap := make(map[string]interface{})
	for k, v := range telemetry {
		telemetryMap[k] = v
	}
	telemetryMap["machineMachineId"] = machineId
	telemetryMap["macMachineId"] = machineId
	telemetryMap["winMachineId"] = machineId
	telemetryMap["linuxMachineId"] = machineId

	newConfig["telemetry"] = telemetryMap
	return true, newConfig
}

// generateMachineId 生成一个新的机器 ID
func generateMachineId() string {
	// 这里使用简单的时间戳作为示例
	// 实际实现应该使用更复杂的算法生成唯一 ID
	return fmt.Sprintf("machine-%d", os.Getpid())
}

// printMap 递归打印 map 结构
func printMap(m map[string]interface{}, level int) {
	indent := strings.Repeat("  ", level)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", indent, k)
			printMap(val, level+1)
		default:
			fmt.Printf("%s%s: %v\n", indent, k, v)
		}
	}
}
