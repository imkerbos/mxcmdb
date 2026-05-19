package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// 构造一个模拟的 asciicast v2 录制数据
func generateRecording(eventCount int) string {
	var b strings.Builder
	header, _ := json.Marshal(map[string]any{
		"version":   2,
		"width":     120,
		"height":    40,
		"timestamp": 1716100000,
		"env":       map[string]string{"SHELL": "/bin/bash", "TERM": "xterm-256color"},
	})
	b.Write(header)
	b.WriteByte('\n')

	for i := 0; i < eventCount; i++ {
		elapsed := float64(i) * 0.05
		data := fmt.Sprintf("output line %d: some terminal data here\r\n", i)
		entry, _ := json.Marshal([]any{elapsed, "o", data})
		b.Write(entry)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestCompressDecompressRecording(t *testing.T) {
	original := generateRecording(100)

	compressed := compressRecording(original)

	// 验证压缩后有 gz: 前缀
	if !strings.HasPrefix(compressed, "gz:") {
		t.Fatalf("压缩结果缺少 gz: 前缀")
	}

	// 验证压缩比
	ratio := float64(len(compressed)) / float64(len(original))
	t.Logf("原始大小: %d bytes, 压缩后: %d bytes, 压缩比: %.2f%%", len(original), len(compressed), ratio*100)
	if ratio >= 1.0 {
		t.Errorf("压缩无效，压缩后大小 >= 原始大小")
	}

	// 验证解压还原
	decompressed := decompressRecording(compressed)
	if decompressed != original {
		t.Fatalf("解压后数据与原始不一致\n原始前100字符: %s\n解压前100字符: %s",
			original[:100], decompressed[:100])
	}
}

func TestDecompressOldUncompressedData(t *testing.T) {
	// 旧格式（无 gz: 前缀）应原样返回
	oldData := generateRecording(10)

	result := decompressRecording(oldData)
	if result != oldData {
		t.Fatalf("旧格式数据应原样返回")
	}
}

func TestDecompressCorruptedData(t *testing.T) {
	// 损坏的压缩数据应原样返回
	corrupted := "gz:notvalidbase64!!!"
	result := decompressRecording(corrupted)
	if result != corrupted {
		t.Fatalf("损坏数据应原样返回")
	}
}

func TestCompressEmptyRecording(t *testing.T) {
	compressed := compressRecording("")
	decompressed := decompressRecording(compressed)
	if decompressed != "" {
		t.Fatalf("空录制数据压缩/解压不一致")
	}
}

func TestCompressLargeRecording(t *testing.T) {
	// 模拟一个长会话（5000 条事件，类似 10 分钟操作）
	original := generateRecording(5000)
	t.Logf("大录制原始大小: %d bytes (%.1f KB)", len(original), float64(len(original))/1024)

	compressed := compressRecording(original)
	ratio := float64(len(compressed)) / float64(len(original))
	t.Logf("压缩后: %d bytes (%.1f KB), 压缩比: %.2f%%", len(compressed), float64(len(compressed))/1024, ratio*100)

	decompressed := decompressRecording(compressed)
	if decompressed != original {
		t.Fatalf("大录制解压不一致")
	}
}

func BenchmarkCompressRecording(b *testing.B) {
	recording := generateRecording(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compressRecording(recording)
	}
}

func BenchmarkDecompressRecording(b *testing.B) {
	recording := generateRecording(1000)
	compressed := compressRecording(recording)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decompressRecording(compressed)
	}
}
