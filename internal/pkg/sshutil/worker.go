package sshutil

import (
	"sync"
)

// Task 执行任务
type Task struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string
}

// TaskResult 任务执行结果
type TaskResult struct {
	Task   Task
	Result ExecResult
}

// WorkerPool 并发执行 SSH 任务
type WorkerPool struct {
	maxConcurrent int
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxConcurrent int) *WorkerPool {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &WorkerPool{maxConcurrent: maxConcurrent}
}

// TaskFunc 自定义任务函数
type TaskFunc func(task Task) ExecResult

// Execute 并发执行任务
func (p *WorkerPool) Execute(tasks []Task, fn TaskFunc) []TaskResult {
	results := make([]TaskResult, len(tasks))
	sem := make(chan struct{}, p.maxConcurrent)
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := fn(t)
			results[idx] = TaskResult{Task: t, Result: result}
		}(i, task)
	}

	wg.Wait()
	return results
}
