package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.emercury.dev/emercury/senderscore/api/internal/data"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure/senderscore"
	"git.emercury.dev/emercury/senderscore/api/internal/usecase"
	"git.emercury.dev/emercury/senderscore/api/pkg/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ip string

func init() {
	parseCmd.Flags().StringVarP(&ip, "ip", "i", "", "IP address to lookup (optional, uses oldest IP if not specified)")
}

var parseCmd = cobra.Command{
	Use:   "parse",
	Short: "Parse sender score data for an IP address",
	Long:  "Fetches sender score data, updates the database, and displays results. If no IP is specified, processes the oldest IP in the database.",
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

		targetIP := ip
		var oldestIP *usecase.IPDTO

		if targetIP == "" {
			oldestIP, err = ipUC.GetOldestIP(ctx)
			if err != nil {
				logrus.WithError(err).Fatal("Failed to get oldest IP from database")
			}
			targetIP = oldestIP.IP

			logrus.WithFields(logrus.Fields{
				"ip":         targetIP,
				"updated_at": time.Unix(oldestIP.UpdatedAt, 0).Format("02.01.2006 15:04:05"),
			}).Info("Processing oldest IP from database")
		} else {
			logrus.WithField("ip", targetIP).Info("Processing specified IP")
		}

		client := &http.Client{Timeout: 15 * time.Second}
		req := senderscore.NewRequestWrapper(client)
		senderClient := senderscore.NewSenderClient(req)
		report, err := senderClient.GetReport(ip)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get sender score report")
		}

		parser := infrastructure.NewParser(report)
		result := parser.Parse()

		logrus.WithFields(logrus.Fields{
			"ip":          targetIP,
			"score":       result.SenderScore,
			"spam_traps":  result.SpamTrap,
			"blocklists":  result.Blocklists,
			"complaints":  result.Complaints,
			"history_cnt": len(result.SSTrend),
		}).Info("Parsed sender score data")

		submitDTO := usecase.SubmitScoreDTO{
			IP:         targetIP,
			Score:      result.SenderScore,
			SpamTrap:   result.SpamTrap,
			Blocklists: result.Blocklists,
			Complaints: result.Complaints,
			History:    convertToHistoryDTO(result),
		}

		submitResult, err := ipUC.SubmitScore(ctx, submitDTO)
		if err != nil {
			logrus.WithError(err).Error("Failed to save to database")
		} else {
			logrus.WithFields(logrus.Fields{
				"ip":              targetIP,
				"ip_created":      submitResult.IPCreated,
				"history_added":   submitResult.HistoryAdded,
				"history_updated": submitResult.HistoryUpdated,
			}).Info("Successfully updated database")
		}

		if oldestIP != nil || submitResult != nil {
			updatedIP, err := ipRepo.GetByIP(ctx, targetIP)
			if err != nil {
				logrus.WithError(err).Warn("Failed to get updated IP for counter update")
			} else {
				for _, groupID := range updatedIP.GroupIDs {
					if err := groupRepo.UpdateCounters(ctx, groupID); err != nil {
						logrus.WithError(err).WithField("group_id", groupID).Warn("Failed to update group counters")
					}
				}
			}
		}

		volumes := make(map[string]int)
		for _, v := range result.SSVolume {
			volumes[v.Timestamp] = v.Value
		}

		displayResults(targetIP, result, volumes)

		logrus.Info("Parse process completed successfully")
	},
}

func displayResults(ip string, result *infrastructure.Result, volumes map[string]int) {
	purple := lipgloss.Color("#7D56F4")
	green := lipgloss.Color("#00C853")
	yellow := lipgloss.Color("#FFD600")
	red := lipgloss.Color("#FF1744")
	baseStyle := lipgloss.NewStyle().Padding(0, 1)

	var scoreColor lipgloss.Color
	switch {
	case result.SenderScore >= 80:
		scoreColor = green
	case result.SenderScore >= 60:
		scoreColor = yellow
	default:
		scoreColor = red
	}

	summaryTable := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 1 && row == 1 { // Score value
				return baseStyle.Foreground(scoreColor).Bold(true)
			}
			return baseStyle
		}).
		Headers("Metric", "Value").
		Rows(
			[]string{"Sender Score", strconv.Itoa(result.SenderScore)},
			[]string{"Spam Traps", strconv.Itoa(result.SpamTrap)},
			[]string{"Blocklists", result.Blocklists},
			[]string{"Complaints", result.Complaints},
		)

	detailTable := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			return baseStyle
		}).
		Headers("DATE", "SCORE", "VOLUME")

	for _, p := range result.SSTrend {
		ms, _ := strconv.ParseInt(p.Timestamp, 10, 64)
		tm := time.Unix(0, ms*int64(time.Millisecond))
		vol := volumes[p.Timestamp]

		detailTable.Row(
			tm.Format("02.01.2006"),
			strconv.Itoa(p.Value),
			strconv.Itoa(vol),
		)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(purple).
		Margin(1, 0)

	historyTitle := lipgloss.NewStyle().
		Margin(1, 0, 0, 1).
		Italic(true).
		Foreground(lipgloss.Color("#888888"))

	reportTitle := fmt.Sprintf("📊 Sender Score Report: %s", ip)

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(reportTitle),
		summaryTable.String(),
		historyTitle.Render(fmt.Sprintf("History (%d days):", len(result.SSTrend))),
		detailTable.String(),
	)

	fmt.Println(ui)
}
