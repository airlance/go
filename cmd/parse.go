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
	parseCmd.Flags().StringVarP(&ip, "ip", "i", "", "IP address to lookup (required)")
	parseCmd.MarkFlagRequired("ip")
}

var parseCmd = cobra.Command{
	Use:   "parse",
	Short: "Run parse process",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		client := &http.Client{Timeout: 15 * time.Second}
		req := senderscore.NewRequestWrapper(client)
		senderClient := senderscore.NewSenderClient(req)
		report, err := senderClient.GetReport(ip)
		if err != nil {
			return
		}

		parser := infrastructure.NewParser(report)
		result := parser.Parse()

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
		submitDTO := usecase.SubmitScoreDTO{
			IP:         ip,
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
				"ip_created":      submitResult.IPCreated,
				"history_added":   submitResult.HistoryAdded,
				"history_updated": submitResult.HistoryUpdated,
			}).Info("Successfully saved to database")
		}

		purple := lipgloss.Color("#7D56F4")
		baseStyle := lipgloss.NewStyle().Padding(0, 1)

		summaryTable := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(purple)).
			StyleFunc(func(row, col int) lipgloss.Style {
				return baseStyle
			}).
			Headers("Name", "Value").
			Rows(
				[]string{"Sender Score", strconv.Itoa(result.SenderScore)},
				[]string{"Spam Traps", strconv.Itoa(result.SpamTrap)},
				[]string{"Blocklists", result.Blocklists},
				[]string{"Complaints", result.Complaints},
			)

		volumes := make(map[string]int)
		for _, v := range result.SSVolume {
			volumes[v.Timestamp] = v.Value
		}

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

		historyTableName := fmt.Sprintf("📊 Report: %s", ip)

		ui := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Margin(1, 0).Render(historyTableName),
			summaryTable.String(),
			lipgloss.NewStyle().Margin(1, 0, 0, 1).Italic(true).Render("History:"),
			detailTable.String(),
		)

		fmt.Println(ui)
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
