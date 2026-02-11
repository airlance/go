package cmd

import (
	"net/http"
	"strconv"
	"time"

	"git.emercury.dev/emercury/senderscore/api/internal/data"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure/senderscore"
	"git.emercury.dev/emercury/senderscore/api/internal/usecase"
	"git.emercury.dev/emercury/senderscore/api/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateCmd = cobra.Command{
	Use:   "update",
	Short: "Update score for the oldest IP address",
	Long:  "Fetches the oldest IP, retrieves current sender score data, and updates the database",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		db, err := infrastructure.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}

		sqlDB, _ := db.DB()
		defer sqlDB.Close()

		groupRepo := data.NewGroupRepository(db)
		ipRepo := data.NewIPRepository(db)
		historyRepo := data.NewHistoryRepository(db)

		ipUC := usecase.NewIPUseCase(groupRepo, ipRepo, historyRepo)
		oldestIP, err := ipUC.GetOldestIP(ctx)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get oldest IP")
		}

		logrus.WithFields(logrus.Fields{
			"ip":         oldestIP.IP,
			"updated_at": time.Unix(oldestIP.UpdatedAt, 0).Format("02.01.2006 15:04:05"),
		}).Info("Processing oldest IP")

		client := &http.Client{Timeout: 15 * time.Second}
		req := senderscore.NewRequestWrapper(client)
		senderClient := senderscore.NewSenderClient(req)

		report, err := senderClient.GetReport(oldestIP.IP)
		if err != nil {
			logrus.WithError(err).Error("Failed to get sender score report")
			return
		}

		parser := infrastructure.NewParser(report)
		result := parser.Parse()

		logrus.WithFields(logrus.Fields{
			"ip":          oldestIP.IP,
			"score":       result.SenderScore,
			"spam_traps":  result.SpamTrap,
			"blocklists":  result.Blocklists,
			"complaints":  result.Complaints,
			"history_cnt": len(result.SSTrend),
		}).Info("Parsed sender score data")

		submitDTO := usecase.SubmitScoreDTO{
			IP:         oldestIP.IP,
			Score:      result.SenderScore,
			SpamTrap:   result.SpamTrap,
			Blocklists: result.Blocklists,
			Complaints: result.Complaints,
			History:    convertToHistoryDTO(result),
		}

		submitResult, err := ipUC.SubmitScore(ctx, submitDTO)
		if err != nil {
			logrus.WithError(err).Error("Failed to submit score")
			return
		}

		logrus.WithFields(logrus.Fields{
			"ip":              oldestIP.IP,
			"ip_created":      submitResult.IPCreated,
			"history_added":   submitResult.HistoryAdded,
			"history_updated": submitResult.HistoryUpdated,
		}).Info("Successfully updated IP score")

		updatedIP, err := ipRepo.GetByIP(ctx, oldestIP.IP)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get updated IP for counter update")
			return
		}

		for _, groupID := range updatedIP.GroupIDs {
			if err := groupRepo.UpdateCounters(ctx, groupID); err != nil {
				logrus.WithError(err).WithField("group_id", groupID).Warn("Failed to update group counters")
			}
		}

		logrus.Info("Update process completed successfully")
	},
}

func convertToHistoryDTO(result *infrastructure.Result) []usecase.HistoryEntryDTO {
	volumes := make(map[string]int)
	for _, v := range result.SSVolume {
		volumes[v.Timestamp] = v.Value
	}

	history := make([]usecase.HistoryEntryDTO, 0, len(result.SSTrend))
	for _, p := range result.SSTrend {
		ms, _ := strconv.ParseInt(p.Timestamp, 10, 64)
		tm := time.Unix(0, ms*int64(time.Millisecond))

		history = append(history, usecase.HistoryEntryDTO{
			Date:     tm.Format("02.01.2006"),
			Score:    p.Value,
			Volume:   volumes[p.Timestamp],
			SpamTrap: result.SpamTrap,
		})
	}

	return history
}
