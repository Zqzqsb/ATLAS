package logger

import (
	"fmt"
	"sync"
	"time"
)

// Logger 进度日志记录器
type Logger struct {
	mu             sync.Mutex
	totalTasks     int
	completedTasks int
	startTime      time.Time
	currentPhase   string
	taskDetails    map[string]*TaskProgress
}

// TaskProgress 任务进度
type TaskProgress struct {
	Name      string
	Status    string // "pending", "running", "completed", "failed"
	StartTime time.Time
	EndTime   time.Time
	Error     string
}

// NewLogger 创建新的日志记录器
func NewLogger(totalTasks int) *Logger {
	return &Logger{
		totalTasks:  totalTasks,
		startTime:   time.Now(),
		taskDetails: make(map[string]*TaskProgress),
	}
}

// SetPhase 设置当前阶段
func (l *Logger) SetPhase(phase string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentPhase = phase
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📍 %s\n", phase)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}

// StartTask 开始任务
func (l *Logger) StartTask(taskName string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.taskDetails[taskName] = &TaskProgress{
		Name:      taskName,
		Status:    "running",
		StartTime: time.Now(),
	}

	fmt.Printf("[%s] 🔄 Started\n", taskName)
}

// CompleteTask 完成任务
func (l *Logger) CompleteTask(taskName string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if task, ok := l.taskDetails[taskName]; ok {
		task.Status = "completed"
		task.EndTime = time.Now()
		l.completedTasks++

		duration := task.EndTime.Sub(task.StartTime)
		fmt.Printf("[%s] ✓ Completed (%.2fs)\n", taskName, duration.Seconds())
		l.printProgress()
	}
}

// FailTask 任务失败
func (l *Logger) FailTask(taskName string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if task, ok := l.taskDetails[taskName]; ok {
		task.Status = "failed"
		task.EndTime = time.Now()
		task.Error = err.Error()
		l.completedTasks++

		fmt.Printf("[%s] ✗ Failed: %v\n", taskName, err)
		l.printProgress()
	}
}

// printProgress 打印进度（内部使用，已加锁）
func (l *Logger) printProgress() {
	if l.totalTasks == 0 {
		return
	}

	percentage := float64(l.completedTasks) / float64(l.totalTasks) * 100
	elapsed := time.Since(l.startTime)

	// 估算剩余时间
	var eta time.Duration
	if l.completedTasks > 0 {
		avgTime := elapsed / time.Duration(l.completedTasks)
		remaining := l.totalTasks - l.completedTasks
		eta = avgTime * time.Duration(remaining)
	}

	fmt.Printf("📊 Progress: %d/%d (%.1f%%) | Elapsed: %s | ETA: %s\n\n",
		l.completedTasks, l.totalTasks, percentage,
		formatDuration(elapsed), formatDuration(eta))
}

// PrintSummary 打印最终摘要
func (l *Logger) PrintSummary() {
	l.mu.Lock()
	defer l.mu.Unlock()

	totalDuration := time.Since(l.startTime)

	var completed, failed int
	for _, task := range l.taskDetails {
		if task.Status == "completed" {
			completed++
		} else if task.Status == "failed" {
			failed++
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📊 Final Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Printf("Total Tasks: %d\n", l.totalTasks)
	fmt.Printf("✓ Completed: %d\n", completed)
	fmt.Printf("✗ Failed: %d\n", failed)
	fmt.Printf("⏱️  Total Time: %s\n", formatDuration(totalDuration))

	if completed > 0 {
		avgTime := totalDuration / time.Duration(completed)
		fmt.Printf("⚡ Avg Time/Task: %s\n", formatDuration(avgTime))
	}

	if failed > 0 {
		fmt.Printf("\n❌ Failed Tasks:\n")
		for _, task := range l.taskDetails {
			if task.Status == "failed" {
				fmt.Printf("  - %s: %s\n", task.Name, task.Error)
			}
		}
	}

	fmt.Printf("\n")
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}

	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// Info 打印信息
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("ℹ️  "+format+"\n", args...)
}

// Warn 打印警告
func (l *Logger) Warn(format string, args ...interface{}) {
	fmt.Printf("⚠️  "+format+"\n", args...)
}

// Error 打印错误
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Printf("❌ "+format+"\n", args...)
}
