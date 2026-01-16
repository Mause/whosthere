package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/core/logging"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run network scanners standalone for debugging/experimentation",
	Long: `Run one or more scanners directly (mdns, ssdp, arp).

Examples:
  whosthere scan -s mdns
  whosthere scan -s "arp,ssdp" -t 30
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scannerNames, _ := cmd.Flags().GetString("scanner")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		scanDuration := time.Duration(timeoutSec) * time.Second

		level := logging.LevelFromEnv(zapcore.InfoLevel)
		_, _, err := logging.Init(level, true)
		if err != nil {
			return err
		}

		ctx := context.Background()

		ouiDB, err := oui.Init(ctx)
		if err != nil {
			zap.L().Warn("failed to initialize OUI DB; continuing without OUI", zap.Error(err))
			ouiDB = nil
		}

		var scList []discovery.Scanner
		sweeper := arp.NewSweeper(5*time.Minute, time.Minute)
		requested := strings.Split(scannerNames, ",")
		for _, r := range requested {
			r = strings.TrimSpace(strings.ToLower(r))
			switch r {
			case "", "all":
				// add all
				scList = append(scList, &ssdp.Scanner{}, arp.NewScanner(sweeper), &mdns.Scanner{})
			case "ssdp":
				scList = append(scList, &ssdp.Scanner{})
			case "arp":
				scList = append(scList, arp.NewScanner(sweeper))
			case "mdns":
				scList = append(scList, &mdns.Scanner{})
			default:
				return fmt.Errorf("unknown scanner: %s", r)
			}
		}

		if len(scList) == 0 {
			return fmt.Errorf("no scanners selected")
		}

		eng := discovery.NewEngine(scList, discovery.WithTimeout(scanDuration), discovery.WithOUIRegistry(ouiDB))

		ctx, cancel := context.WithTimeout(ctx, scanDuration)
		defer cancel()

		devices, err := eng.Stream(ctx, func(d discovery.Device) {
		})
		if err != nil {
			return err
		}

		zap.L().Info("scan complete", zap.Int("devices", len(devices)))
		for _, d := range devices {
			zap.L().Info("device",
				zap.String("ip", d.IP.String()),
				zap.String("hostname", d.DisplayName),
				zap.String("mac", d.MAC),
				zap.String("manufacturer", d.Manufacturer),
			)
		}
		return nil
	},
}

func init() {
	scanCmd.Flags().StringP("scanner", "s", "all", "Comma-separated scanners to run (mdns,ssdp,arp,all)")
	scanCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the scan")
	rootCmd.AddCommand(scanCmd)
}
