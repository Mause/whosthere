package cmd

import (
	"context"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/logging"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type InitResult struct {
	Logger    *zap.Logger
	LogPath   string
	Config    *config.Config
	OuiDB     *oui.Registry
	Interface *discovery.InterfaceInfo
}

func InitComponents(configFileOverride string, enableStdout bool) (*InitResult, error) {
	level := logging.LevelFromEnv(zapcore.InfoLevel)
	logger, logPath, err := logging.Init(level, enableStdout)
	if err != nil {
		return nil, err
	}

	apiLevel := androidDeviceApiLevel()

	logger.Info("android api level detected", zap.Int("api_level", apiLevel))

	cfg, err := config.Load(configFileOverride)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	ouiDB, err := oui.Init(ctx)
	if err != nil {
		zap.L().Warn("failed to initialize OUI DB; continuing without OUI", zap.Error(err))
		ouiDB = nil
	}

	iface, err := discovery.NewInterfaceInfo(cfg.NetworkInterface)
	if err != nil {
		return nil, err
	}

	return &InitResult{
		Logger:    logger,
		LogPath:   logPath,
		Config:    cfg,
		OuiDB:     ouiDB,
		Interface: iface,
	}, nil
}
