package repository

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// getTestDB 获取测试数据库连接（dev 环境 PostgreSQL）
func getTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=mxcmdb password=mxcmdb123 dbname=mxcmdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Skipf("跳过数据库测试（连接失败）: %v", err)
	}

	// 配置连接池（模拟生产环境）
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db
}

// seedAssets 插入 n 条测试资产，返回 IDs（用完后调用 cleanupAssets 清理）
func seedAssets(t *testing.T, db *gorm.DB, n int) []uint {
	assets := make([]model.Asset, n)
	for i := range assets {
		assets[i] = model.Asset{
			Hostname:    fmt.Sprintf("bench-%04d", i+1),
			IP:          fmt.Sprintf("172.30.%d.%d", (i/256)%256, i%256),
			Port:        22,
			SshUser:     "root",
			Type:        "server",
			Source:      "manual",
			Status:      "unknown",
			Environment: "bench",
		}
	}
	if err := db.CreateInBatches(assets, 200).Error; err != nil {
		t.Fatalf("seed 资产失败: %v", err)
	}
	ids := make([]uint, n)
	for i := range assets {
		ids[i] = assets[i].ID
	}
	return ids
}

func cleanupAssets(db *gorm.DB, ids []uint) {
	db.Where("id IN ?", ids).Unscoped().Delete(&model.Asset{})
}

// ===================================================================
//  1. 逐条 vs 批量查询 —— 真实数据库对比
// ===================================================================

func TestGetByIDs_vs_GetByID_1000(t *testing.T) {
	db := getTestDB(t)
	ids := seedAssets(t, db, 1000)
	defer cleanupAssets(db, ids)

	repo := NewAssetRepository(db)

	// 批量查询
	start := time.Now()
	batchResult, err := repo.GetByIDs(ids)
	batchTime := time.Since(start)
	if err != nil {
		t.Fatalf("GetByIDs 失败: %v", err)
	}

	// 逐条查询（取 100 条采样，线性外推）
	sampleN := 100
	start = time.Now()
	for _, id := range ids[:sampleN] {
		_, _ = repo.GetByID(id)
	}
	singleSample := time.Since(start)
	singleEstimate := time.Duration(float64(singleSample) / float64(sampleN) * 1000)

	t.Logf("━━━ GetByIDs vs GetByID (1000 台) ━━━")
	t.Logf("  GetByIDs:       %v (%d 条)", batchTime, len(batchResult))
	t.Logf("  GetByID 采样:   %v / %d 条 → 推算 1000 条: %v", singleSample, sampleN, singleEstimate)
	t.Logf("  加速比:         %.0fx", float64(singleEstimate)/float64(batchTime))

	if len(batchResult) != 1000 {
		t.Errorf("结果数不对: %d", len(batchResult))
	}
}

// ===================================================================
//  2. 逐条 vs 批量写入 —— 真实数据库对比
// ===================================================================

func TestCreateInBatches_vs_Create_1000(t *testing.T) {
	db := getTestDB(t)

	task1 := &model.Task{Name: "bench_single_1000", Type: "command", Command: "echo 1", Status: "pending"}
	task2 := &model.Task{Name: "bench_batch_1000", Type: "command", Command: "echo 2", Status: "pending"}
	db.Create(task1)
	db.Create(task2)
	defer func() {
		db.Where("task_id IN ?", []uint{task1.ID, task2.ID}).Unscoped().Delete(&model.TaskResult{})
		db.Unscoped().Delete(&model.Task{}, task1.ID)
		db.Unscoped().Delete(&model.Task{}, task2.ID)
	}()

	n := 1000

	// 逐条插入
	start := time.Now()
	for i := 0; i < n; i++ {
		db.Create(&model.TaskResult{
			TaskID:   task1.ID,
			AssetID:  uint(i + 1),
			Hostname: fmt.Sprintf("host-%04d", i+1),
			IP:       fmt.Sprintf("10.0.%d.%d", (i/256)%256, i%256),
			Status:   "pending",
		})
	}
	singleTime := time.Since(start)

	// 批量插入
	results := make([]*model.TaskResult, n)
	for i := range results {
		results[i] = &model.TaskResult{
			TaskID:   task2.ID,
			AssetID:  uint(i + 1),
			Hostname: fmt.Sprintf("host-%04d", i+1),
			IP:       fmt.Sprintf("10.0.%d.%d", (i/256)%256, i%256),
			Status:   "pending",
		}
	}
	start = time.Now()
	db.CreateInBatches(results, 100)
	batchTime := time.Since(start)

	// 验证 ID 回填
	allHaveID := true
	for _, r := range results {
		if r.ID == 0 {
			allHaveID = false
			break
		}
	}

	t.Logf("━━━ CreateInBatches vs Create (%d 条) ━━━", n)
	t.Logf("  逐条插入: %v (%.2f ms/条)", singleTime, float64(singleTime.Microseconds())/float64(n)/1000)
	t.Logf("  批量插入: %v (%.2f ms/条)", batchTime, float64(batchTime.Microseconds())/float64(n)/1000)
	t.Logf("  加速比:   %.1fx", float64(singleTime)/float64(batchTime))
	t.Logf("  ID 回填:  %v", allHaveID)

	// 验证数据完整
	var count1, count2 int64
	db.Model(&model.TaskResult{}).Where("task_id = ?", task1.ID).Count(&count1)
	db.Model(&model.TaskResult{}).Where("task_id = ?", task2.ID).Count(&count2)
	if count1 != int64(n) || count2 != int64(n) {
		t.Errorf("数据不完整: single=%d, batch=%d", count1, count2)
	}
}

// ===================================================================
//  3. 数据库并发写入压力 —— 模拟 10 个任务同时写 1000 台结果
// ===================================================================

func TestConcurrentDBWrite_10Tasks_1000Assets(t *testing.T) {
	db := getTestDB(t)

	taskCount := 10
	assetCount := 1000

	// 创建 10 个任务
	taskIDs := make([]uint, taskCount)
	for i := 0; i < taskCount; i++ {
		task := &model.Task{
			Name:    fmt.Sprintf("concurrent_bench_%d", i),
			Type:    "command",
			Command: "echo test",
			Status:  "pending",
		}
		db.Create(task)
		taskIDs[i] = task.ID
	}
	defer func() {
		for _, id := range taskIDs {
			db.Where("task_id = ?", id).Unscoped().Delete(&model.TaskResult{})
			db.Unscoped().Delete(&model.Task{}, id)
		}
	}()

	// 记录初始内存
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// 并发：10 个 goroutine 各插入 1000 条
	var wg sync.WaitGroup
	var totalErrors atomic.Int64
	start := time.Now()

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(taskID uint) {
			defer wg.Done()

			results := make([]*model.TaskResult, assetCount)
			for j := range results {
				results[j] = &model.TaskResult{
					TaskID:   taskID,
					AssetID:  uint(j + 1),
					Hostname: fmt.Sprintf("host-%04d", j+1),
					IP:       fmt.Sprintf("10.0.%d.%d", (j/256)%256, j%256),
					Status:   "pending",
					Stdout:   fmt.Sprintf("hostname\nCPU: 4 cores\nMem: 8GB\nDisk: 200GB\nOS: Ubuntu 22.04\nKernel: 5.15.0\nUptime: %d days", j%365),
					ExitCode: 0,
					Duration: 200,
				}
			}
			if err := db.CreateInBatches(results, 100).Error; err != nil {
				totalErrors.Add(1)
			}
		}(taskIDs[i])
	}
	wg.Wait()
	elapsed := time.Since(start)

	// 记录写入后内存
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	totalRecords := taskCount * assetCount
	tps := float64(totalRecords) / elapsed.Seconds()

	// 验证数据
	var dbCount int64
	db.Model(&model.TaskResult{}).Where("task_id IN ?", taskIDs).Count(&dbCount)

	t.Logf("━━━ 并发数据库写入压力测试 ━━━")
	t.Logf("  任务数: %d × 资产数: %d = 总记录: %d", taskCount, assetCount, totalRecords)
	t.Logf("  总耗时:     %v", elapsed)
	t.Logf("  写入 TPS:   %.0f 条/秒", tps)
	t.Logf("  错误数:     %d", totalErrors.Load())
	t.Logf("  数据库验证: %d 条 (期望 %d)", dbCount, totalRecords)
	t.Logf("  内存增长:   %.1f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)

	if dbCount != int64(totalRecords) {
		t.Errorf("数据丢失: 期望 %d, 实际 %d", totalRecords, dbCount)
	}
	if totalErrors.Load() > 0 {
		t.Errorf("写入错误: %d", totalErrors.Load())
	}
}

// ===================================================================
//  4. 数据库并发读写混合 —— 模拟写入同时查询
// ===================================================================

func TestConcurrentReadWrite_1000(t *testing.T) {
	db := getTestDB(t)
	ids := seedAssets(t, db, 1000)
	defer cleanupAssets(db, ids)

	repo := NewAssetRepository(db)

	// 场景: 5 个 writer 各写 200 条 + 10 个 reader 持续读 1000 条资产
	task := &model.Task{Name: "rw_bench", Type: "command", Command: "echo rw", Status: "pending"}
	db.Create(task)
	defer func() {
		db.Where("task_id = ?", task.ID).Unscoped().Delete(&model.TaskResult{})
		db.Unscoped().Delete(&model.Task{}, task.ID)
	}()

	var wg sync.WaitGroup
	var readCount, writeCount atomic.Int64
	var readErrors, writeErrors atomic.Int64
	done := make(chan struct{})

	start := time.Now()

	// Writers: 5 goroutines 各批量写 200 条
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			results := make([]*model.TaskResult, 200)
			for j := range results {
				results[j] = &model.TaskResult{
					TaskID:   task.ID,
					AssetID:  uint(workerID*200 + j + 1),
					Hostname: fmt.Sprintf("host-%d-%d", workerID, j),
					IP:       "10.0.0.1",
					Status:   "success",
					Stdout:   "ok",
				}
			}
			if err := db.CreateInBatches(results, 50).Error; err != nil {
				writeErrors.Add(1)
			} else {
				writeCount.Add(int64(len(results)))
			}
		}(w)
	}

	// Readers: 10 goroutines 持续 GetByIDs
	for r := 0; r < 10; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					result, err := repo.GetByIDs(ids)
					if err != nil {
						readErrors.Add(1)
					} else {
						readCount.Add(int64(len(result)))
					}
				}
			}
		}()
	}

	// 等 writer 写完
	time.Sleep(1 * time.Second) // 确保 reader 跑起来
	close(done)
	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("━━━ 并发读写混合测试 ━━━")
	t.Logf("  Writer: 5 × 200 条 = %d 条 (错误 %d)", writeCount.Load(), writeErrors.Load())
	t.Logf("  Reader: 10 并发 × GetByIDs(1000) = %d 条总读取 (错误 %d)", readCount.Load(), readErrors.Load())
	t.Logf("  总耗时: %v", elapsed)
	t.Logf("  读 QPS: %.0f 次/秒 (每次 1000 条)", float64(readCount.Load())/1000/elapsed.Seconds())

	if writeErrors.Load() > 0 {
		t.Errorf("写入错误: %d", writeErrors.Load())
	}
	if readErrors.Load() > 0 {
		t.Errorf("读取错误: %d", readErrors.Load())
	}
}

// ===================================================================
//  5. 连接池耗尽测试 —— 并发数超过连接池上限
// ===================================================================

func TestConnectionPoolExhaustion(t *testing.T) {
	dsn := "host=localhost port=5432 user=mxcmdb password=mxcmdb123 dbname=mxcmdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Skipf("跳过: %v", err)
	}

	// 故意设置较小的连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)

	ids := seedAssets(t, db, 500)
	defer cleanupAssets(db, ids)

	repo := NewAssetRepository(db)

	// 50 并发查询，但连接池只有 20
	var wg sync.WaitGroup
	var successCount, errorCount atomic.Int64

	start := time.Now()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := repo.GetByIDs(ids)
			if err != nil {
				errorCount.Add(1)
			} else if len(result) == 500 {
				successCount.Add(1)
			} else {
				errorCount.Add(1)
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	stats := sqlDB.Stats()

	t.Logf("━━━ 连接池压力测试 (MaxOpen=20, 50并发) ━━━")
	t.Logf("  成功: %d/50, 错误: %d/50", successCount.Load(), errorCount.Load())
	t.Logf("  总耗时: %v", elapsed)
	t.Logf("  连接池状态:")
	t.Logf("    OpenConnections: %d", stats.OpenConnections)
	t.Logf("    InUse: %d", stats.InUse)
	t.Logf("    Idle: %d", stats.Idle)
	t.Logf("    WaitCount: %d (等待连接次数)", stats.WaitCount)
	t.Logf("    WaitDuration: %v (累计等待时间)", stats.WaitDuration)

	if errorCount.Load() > 0 {
		t.Errorf("有 %d 个查询失败", errorCount.Load())
	}
}

// ===================================================================
//  6. 资产列表分页查询 —— 万级数据下分页性能
// ===================================================================

func TestAssetListPagination_10000(t *testing.T) {
	db := getTestDB(t)
	ids := seedAssets(t, db, 5000)
	defer cleanupAssets(db, ids)

	repo := NewAssetRepository(db)

	// 第 1 页
	start := time.Now()
	list1, total, err := repo.List("", "server", "manual", "", "bench", "", 0, "", 1, 20)
	p1Time := time.Since(start)
	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}

	// 中间页
	midPage := int(total) / 20 / 2
	if midPage < 1 {
		midPage = 1
	}
	start = time.Now()
	listMid, _, _ := repo.List("", "server", "manual", "", "bench", "", 0, "", midPage, 20)
	midTime := time.Since(start)

	// 末页
	lastPage := int(total) / 20
	if lastPage < 1 {
		lastPage = 1
	}
	start = time.Now()
	listLast, _, _ := repo.List("", "server", "manual", "", "bench", "", 0, "", lastPage, 20)
	lastTime := time.Since(start)

	// 带关键字搜索
	start = time.Now()
	listSearch, searchTotal, _ := repo.List("bench-0042", "server", "", "", "", "", 0, "", 1, 20)
	searchTime := time.Since(start)

	t.Logf("━━━ 资产列表分页查询 (5000 条基础数据) ━━━")
	t.Logf("  总数: %d", total)
	t.Logf("  第 1 页:   %v (%d 条)", p1Time, len(list1))
	t.Logf("  第 %d 页:  %v (%d 条)", midPage, midTime, len(listMid))
	t.Logf("  末页 %d:   %v (%d 条)", lastPage, lastTime, len(listLast))
	t.Logf("  关键字搜索: %v (%d 条, 匹配 %d)", searchTime, len(listSearch), searchTotal)

	if p1Time > 500*time.Millisecond {
		t.Errorf("首页查询过慢: %v", p1Time)
	}
	if lastTime > 500*time.Millisecond {
		t.Errorf("末页查询过慢: %v", lastTime)
	}
}
