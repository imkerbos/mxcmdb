package service

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
)

// ===================================================================
//  1000 台主机压力测试 —— 应用层
//  重点: Worker Pool 真实行为、内存/goroutine 监控、混合负载
// ===================================================================

func generateAssets(n int) []model.Asset {
	assets := make([]model.Asset, n)
	for i := range assets {
		assets[i] = model.Asset{
			BaseModel: model.BaseModel{ID: uint(i + 1)},
			Hostname:  fmt.Sprintf("host-%04d", i+1),
			IP:        fmt.Sprintf("10.%d.%d.%d", (i/65536)%256, (i/256)%256, i%256),
			Port:      22,
			SshUser:   "root",
		}
	}
	return assets
}

func buildAssetMap(assets []model.Asset) map[uint]*model.Asset {
	m := make(map[uint]*model.Asset, len(assets))
	for i := range assets {
		m[assets[i].ID] = &assets[i]
	}
	return m
}

func generateSSHTasks(n int) []sshutil.Task {
	tasks := make([]sshutil.Task, n)
	for i := range tasks {
		tasks[i] = sshutil.Task{
			Host:     fmt.Sprintf("10.0.%d.%d", (i/256)%256, i%256),
			Port:     22,
			User:     "root",
			Password: "testpassword",
		}
	}
	return tasks
}

func generateLongRecording(eventCount int) string {
	var b strings.Builder
	header, _ := json.Marshal(map[string]any{
		"version":   2,
		"width":     120,
		"height":    40,
		"timestamp": time.Now().Unix(),
		"env":       map[string]string{"SHELL": "/bin/bash", "TERM": "xterm-256color"},
	})
	b.Write(header)
	b.WriteByte('\n')

	for i := 0; i < eventCount; i++ {
		elapsed := float64(i) * 0.05
		data := fmt.Sprintf("\033[32mroot@host-%04d\033[0m:\033[34m~\033[0m# ls -la /var/log/\r\ntotal %d\r\n-rw-r--r-- 1 root root %d May 19 10:00 syslog\r\n", i%1000, i*100, i*1024)
		entry, _ := json.Marshal([]any{elapsed, "o", data})
		b.Write(entry)
		b.WriteByte('\n')
	}
	return b.String()
}

func getMemStats() runtime.MemStats {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// ===================================================================
//  1. Worker Pool — goroutine 泄漏检测
//     确认 Execute 返回后没有残留 goroutine
// ===================================================================

func TestWorkerPool_GoroutineLeak_1000(t *testing.T) {
	goroutinesBefore := runtime.NumGoroutine()

	tasks := generateSSHTasks(1000)
	pool := sshutil.NewWorkerPool(100)

	pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		time.Sleep(10 * time.Millisecond)
		return sshutil.ExecResult{ExitCode: 0}
	})

	// 给 runtime 一点时间回收
	time.Sleep(50 * time.Millisecond)
	goroutinesAfter := runtime.NumGoroutine()

	leaked := goroutinesAfter - goroutinesBefore
	t.Logf("━━━ Goroutine 泄漏检测 ━━━")
	t.Logf("  Execute 前: %d goroutines", goroutinesBefore)
	t.Logf("  Execute 后: %d goroutines", goroutinesAfter)
	t.Logf("  残留: %d", leaked)

	if leaked > 5 { // 允许少量 runtime 内部 goroutine 波动
		t.Errorf("可能存在 goroutine 泄漏: +%d", leaked)
	}
}

// ===================================================================
//  2. Worker Pool — 内存占用与分配
//     1000 任务完成后的堆内存增量
// ===================================================================

func TestWorkerPool_MemoryUsage_1000(t *testing.T) {
	memBefore := getMemStats()

	tasks := generateSSHTasks(1000)
	pool := sshutil.NewWorkerPool(100)

	results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		time.Sleep(5 * time.Millisecond)
		// 模拟真实 SSH 输出（每台约 2KB 返回数据）
		return sshutil.ExecResult{
			Stdout: strings.Repeat("a]", 1024),
			ExitCode: 0,
			Duration: 5 * time.Millisecond,
		}
	})

	memAfter := getMemStats()
	heapGrowth := float64(memAfter.HeapInuse-memBefore.HeapInuse) / 1024 / 1024
	totalAlloc := float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / 1024 / 1024

	t.Logf("━━━ 1000 台主机内存分析 ━━━")
	t.Logf("  结果数:        %d", len(results))
	t.Logf("  堆内存增长:    %.2f MB", heapGrowth)
	t.Logf("  总分配:        %.2f MB", totalAlloc)
	t.Logf("  每台平均:      %.1f KB", totalAlloc*1024/1000)
	t.Logf("  GC 次数:       %d → %d (+%d)", memBefore.NumGC, memAfter.NumGC, memAfter.NumGC-memBefore.NumGC)

	if heapGrowth > 100 {
		t.Errorf("堆内存增长异常: %.2f MB (期望 < 100MB)", heapGrowth)
	}
}

// ===================================================================
//  3. Worker Pool — 并发上限与调度效率
//     验证信号量实际控制了并发数
// ===================================================================

func TestWorkerPool_ConcurrencyControl(t *testing.T) {
	tasks := generateSSHTasks(500)

	testCases := []struct {
		maxConcurrent int
		taskDuration  time.Duration
	}{
		{10, 50 * time.Millisecond},
		{50, 50 * time.Millisecond},
		{100, 50 * time.Millisecond},
		{200, 50 * time.Millisecond},
	}

	t.Logf("━━━ 并发控制验证 (500 任务, 50ms/个) ━━━")
	t.Logf("  %-15s %-12s %-12s %-8s", "并发上限", "实际耗时", "理论耗时", "效率")

	for _, tc := range testCases {
		pool := sshutil.NewWorkerPool(tc.maxConcurrent)

		var peakConcurrent atomic.Int64
		var currentConcurrent atomic.Int64

		start := time.Now()
		pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
			cur := currentConcurrent.Add(1)
			// 记录峰值并发
			for {
				peak := peakConcurrent.Load()
				if cur <= peak || peakConcurrent.CompareAndSwap(peak, cur) {
					break
				}
			}
			time.Sleep(tc.taskDuration)
			currentConcurrent.Add(-1)
			return sshutil.ExecResult{ExitCode: 0}
		})
		elapsed := time.Since(start)

		theoreticalMin := time.Duration(float64(len(tasks)) / float64(tc.maxConcurrent) * float64(tc.taskDuration))
		efficiency := float64(theoreticalMin) / float64(elapsed) * 100

		t.Logf("  %-15d %-12v %-12v %.1f%%  (峰值并发: %d)",
			tc.maxConcurrent, elapsed.Round(time.Millisecond), theoreticalMin, efficiency, peakConcurrent.Load())

		// 峰值并发不应超过设定上限
		if int(peakConcurrent.Load()) > tc.maxConcurrent {
			t.Errorf("并发超限: peak=%d > max=%d", peakConcurrent.Load(), tc.maxConcurrent)
		}
		// 效率不应低于 70%
		if efficiency < 70 {
			t.Errorf("调度效率过低: %.1f%%", efficiency)
		}
	}
}

// ===================================================================
//  4. Worker Pool — 部分失败场景
//     模拟 10% 超时 + 5% 连接失败
// ===================================================================

func TestWorkerPool_PartialFailure_1000(t *testing.T) {
	tasks := generateSSHTasks(1000)
	pool := sshutil.NewWorkerPool(100)

	var successCount, timeoutCount, connFailCount atomic.Int64

	start := time.Now()
	results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		hostNum := 0
		fmt.Sscanf(task.Host, "10.0.%d", &hostNum)

		// 10% 超时 (慢响应)
		if hostNum%10 == 0 {
			time.Sleep(500 * time.Millisecond)
			timeoutCount.Add(1)
			return sshutil.ExecResult{
				ExitCode: -1,
				Duration: 500 * time.Millisecond,
				Err:      fmt.Errorf("command timeout"),
			}
		}

		// 5% 连接失败
		if hostNum%20 == 1 {
			connFailCount.Add(1)
			return sshutil.ExecResult{
				ExitCode: -1,
				Err:      fmt.Errorf("connection refused"),
			}
		}

		// 正常
		time.Sleep(50 * time.Millisecond)
		successCount.Add(1)
		return sshutil.ExecResult{
			Stdout:   "ok",
			ExitCode: 0,
			Duration: 50 * time.Millisecond,
		}
	})
	elapsed := time.Since(start)

	// 统计结果
	resultSuccess := 0
	resultFailed := 0
	for _, r := range results {
		if r.Result.Err == nil && r.Result.ExitCode == 0 {
			resultSuccess++
		} else {
			resultFailed++
		}
	}

	t.Logf("━━━ 部分失败场景 (1000 台, 10%%超时 + 5%%连接失败) ━━━")
	t.Logf("  总耗时:     %v", elapsed)
	t.Logf("  结果汇总:   成功 %d, 失败 %d (总 %d)", resultSuccess, resultFailed, len(results))
	t.Logf("  失败明细:   超时 %d, 连接失败 %d", timeoutCount.Load(), connFailCount.Load())

	// 总数必须一致
	if len(results) != 1000 {
		t.Errorf("结果数不一致: %d (期望 1000)", len(results))
	}
	// 成功 + 失败 = 1000
	if resultSuccess+resultFailed != 1000 {
		t.Errorf("计数不一致: %d + %d != 1000", resultSuccess, resultFailed)
	}
}

// ===================================================================
//  5. 混合负载 —— 模拟实际场景: 探针 + 任务 + 录像压缩同时跑
// ===================================================================

func TestMixedWorkload_1000(t *testing.T) {
	memBefore := getMemStats()
	goroutinesBefore := runtime.NumGoroutine()

	var wg sync.WaitGroup
	var probeResults, taskResults, compressionResults atomic.Int64
	var probeFails, taskFails atomic.Int64

	start := time.Now()

	// 负载 A: 探针采集 1000 台 (100 并发, 200ms/台)
	wg.Add(1)
	go func() {
		defer wg.Done()
		tasks := generateSSHTasks(1000)
		pool := sshutil.NewWorkerPool(100)
		results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
			time.Sleep(200 * time.Millisecond)
			return sshutil.ExecResult{
				Stdout: fmt.Sprintf(
					"===HOSTNAME===\n%s\n===CPU===\nCPU(s): 8\n===MEMORY===\nMem: 16384\n===DISK===\n/dev/sda1 500G\n===OS===\nUbuntu 22.04\n===KERNEL===\n5.15.0",
					task.Host,
				),
				ExitCode: 0,
				Duration: 200 * time.Millisecond,
			}
		})
		for _, r := range results {
			if r.Result.Err == nil {
				probeResults.Add(1)
			} else {
				probeFails.Add(1)
			}
		}
	}()

	// 负载 B: 批量命令执行 500 台 (50 并发, 100ms/台)
	wg.Add(1)
	go func() {
		defer wg.Done()
		tasks := generateSSHTasks(500)
		pool := sshutil.NewWorkerPool(50)
		results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
			time.Sleep(100 * time.Millisecond)
			return sshutil.ExecResult{
				Stdout:   "service nginx status: active (running)",
				ExitCode: 0,
				Duration: 100 * time.Millisecond,
			}
		})
		for _, r := range results {
			if r.Result.Err == nil {
				taskResults.Add(1)
			} else {
				taskFails.Add(1)
			}
		}
	}()

	// 负载 C: 20 个终端会话同时断开，并发压缩录像
	wg.Add(1)
	go func() {
		defer wg.Done()
		var innerWg sync.WaitGroup
		for i := 0; i < 20; i++ {
			innerWg.Add(1)
			go func(idx int) {
				defer innerWg.Done()
				recording := generateLongRecording(500 + idx*50)
				compressed := compressRecording(recording)
				decompressed := decompressRecording(compressed)
				if decompressed == recording {
					compressionResults.Add(1)
				}
			}(i)
		}
		innerWg.Wait()
	}()

	wg.Wait()
	elapsed := time.Since(start)

	goroutinesAfter := runtime.NumGoroutine()
	memAfter := getMemStats()

	t.Logf("━━━ 混合负载测试 ━━━")
	t.Logf("  并行任务:")
	t.Logf("    探针采集:  1000 台 / 100 并发 → 成功 %d, 失败 %d", probeResults.Load(), probeFails.Load())
	t.Logf("    命令执行:  500 台 / 50 并发  → 成功 %d, 失败 %d", taskResults.Load(), taskFails.Load())
	t.Logf("    录像压缩:  20 个终端          → 成功 %d", compressionResults.Load())
	t.Logf("  总耗时:     %v", elapsed)
	t.Logf("  Goroutine:  %d → %d (残留 %d)", goroutinesBefore, goroutinesAfter, goroutinesAfter-goroutinesBefore)
	t.Logf("  堆内存增长: %.1f MB", float64(memAfter.HeapInuse-memBefore.HeapInuse)/1024/1024)
	t.Logf("  总分配:     %.1f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)

	if probeResults.Load() != 1000 {
		t.Errorf("探针应全部成功: %d/1000", probeResults.Load())
	}
	if taskResults.Load() != 500 {
		t.Errorf("任务应全部成功: %d/500", taskResults.Load())
	}
	if compressionResults.Load() != 20 {
		t.Errorf("录像压缩应全部成功: %d/20", compressionResults.Load())
	}

	goroutineLeak := goroutinesAfter - goroutinesBefore
	if goroutineLeak > 5 {
		t.Errorf("goroutine 泄漏: +%d", goroutineLeak)
	}
}

// ===================================================================
//  6. 10000 台主机极限测试 — Worker Pool + 内存
// ===================================================================

func TestWorkerPool_10000_Concurrent200(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过极限测试 (-short)")
	}

	memBefore := getMemStats()
	goroutinesBefore := runtime.NumGoroutine()

	tasks := generateSSHTasks(10000)
	pool := sshutil.NewWorkerPool(200)

	var peakGoroutines atomic.Int64

	start := time.Now()
	results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		cur := int64(runtime.NumGoroutine())
		for {
			peak := peakGoroutines.Load()
			if cur <= peak || peakGoroutines.CompareAndSwap(peak, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		return sshutil.ExecResult{
			Stdout:   strings.Repeat("x", 512),
			ExitCode: 0,
			Duration: 50 * time.Millisecond,
		}
	})
	elapsed := time.Since(start)

	time.Sleep(100 * time.Millisecond)
	goroutinesAfter := runtime.NumGoroutine()
	memAfter := getMemStats()

	successCount := 0
	for _, r := range results {
		if r.Result.Err == nil && r.Result.ExitCode == 0 {
			successCount++
		}
	}

	theoreticalMin := time.Duration(float64(10000) / 200 * float64(50*time.Millisecond))
	efficiency := float64(theoreticalMin) / float64(elapsed) * 100

	t.Logf("━━━ 10000 台主机极限测试 ━━━")
	t.Logf("  配置:        10000 任务, 200 并发, 50ms/个")
	t.Logf("  结果:        成功 %d/10000", successCount)
	t.Logf("  总耗时:      %v (理论: %v, 效率: %.1f%%)", elapsed, theoreticalMin, efficiency)
	t.Logf("  吞吐量:      %.0f 台/秒", float64(10000)/elapsed.Seconds())
	t.Logf("  峰值 goroutine: %d", peakGoroutines.Load())
	t.Logf("  残留 goroutine: %d", goroutinesAfter-goroutinesBefore)
	t.Logf("  堆内存增长:  %.1f MB", float64(memAfter.HeapInuse-memBefore.HeapInuse)/1024/1024)
	t.Logf("  总分配:      %.1f MB (%.1f KB/台)", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024, float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/10000)

	if successCount != 10000 {
		t.Errorf("应全部成功")
	}
	if efficiency < 70 {
		t.Errorf("效率过低: %.1f%%", efficiency)
	}
}
