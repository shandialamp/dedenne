package dedenne

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/shandialamp/dedenne/config"
	"github.com/shandialamp/dedenne/log"
	"go.uber.org/zap"
)

func Start(registerRoutes func(e *echo.Echo, cfg *config.Config)) {
	// 读取配置
	cfg, err := config.ReadConfig("")
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

	// 创建 Echo 应用
	e := echo.New()

	// 注册全局中间件
	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())

	// 注册路由
	registerRoutes(e, cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Server listening", zap.String("addr", addr))
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		logger.Error("Server error", zap.Error(err))
	}
}
