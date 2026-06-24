package dedenne

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/shandialamp/dedenne/bizerr"
	"github.com/shandialamp/dedenne/config"
	"github.com/shandialamp/dedenne/database"
	"github.com/shandialamp/dedenne/log"
	"go.uber.org/zap"
)

// StartOption 是 Start 函数的选项函数类型
type StartOption func(*startOptions)

// startOptions 保存 Start 函数的选项
type startOptions struct {
	configPath string
	setup      func(e *echo.Echo, cfg *config.Config, db *sql.DB)
}

// WithConfigPath 设置配置文件路径
func WithConfigPath(path string) StartOption {
	return func(opts *startOptions) {
		opts.configPath = path
	}
}

// WithSetup 设置初始化函数（注册路由、中间件、服务等）
func WithSetup(fn func(e *echo.Echo, cfg *config.Config, db *sql.DB)) StartOption {
	return func(opts *startOptions) {
		opts.setup = fn
	}
}

// Start 启动应用服务器
func Start(opts ...StartOption) {
	// 初始化选项
	options := &startOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 读取配置
	cfg, err := config.ReadConfig(options.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := log.InitLogger(&cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	logger := log.L()
	logger.Info("Application starting",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// 初始化数据库
	db, err := database.NewDB(&cfg.Database, logger)
	if err != nil {
		logger.Error("Failed to initialize database",
			zap.Error(err),
			zap.Stack("stacktrace"),
		)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(db, logger); err != nil {
			logger.Error("Failed to close database",
				zap.Error(err),
				zap.Stack("stacktrace"),
			)
		}
	}()

	// 创建 Echo 应用
	e := echo.New()

	// 注册自定义错误处理器
	e.HTTPErrorHandler = bizerr.HTTPErrorHandler(logger)

	// 注册全局中间件
	// 移除 RequestLogger 以避免与错误处理器冲突（请求日志在错误处理器中记录）
	e.Use(echomiddleware.Recover())

	// 执行初始化
	if options.setup != nil {
		options.setup(e, cfg, db)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Server listening", zap.String("addr", addr))
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		logger.Error("Server error",
			zap.Error(err),
			zap.Stack("stacktrace"),
		)
	}
}
