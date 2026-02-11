package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure/senderscore"
	"git.emercury.dev/emercury/senderscore/api/pkg/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	submitIP       string
	submitAPIURL   string
	submitAPIToken string
)

func init() {
	submitCmd.Flags().StringVarP(&submitIP, "ip", "i", "", "IP address to lookup (required)")
	submitCmd.Flags().StringVarP(&submitAPIURL, "api-url", "u", "http://localhost:8080", "API base URL")
	submitCmd.Flags().StringVarP(&submitAPIToken, "token", "t", "", "API authentication token (required)")
	submitCmd.MarkFlagRequired("ip")
	submitCmd.MarkFlagRequired("token")
}

var submitCmd = cobra.Command{
	Use:   "submit",
	Short: "Parse IP data and submit to API",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		_ = config.Init(ctx)

		// Получение данных
		client := &http.Client{Timeout: 15 * time.Second}
		req := senderscore.NewRequestWrapper(client)
		senderClient := senderscore.NewSenderClient(req)
		report, err := senderClient.GetReport(submitIP)
		if err != nil {
			logrus.WithError(err).Error("Failed to get report")
			return
		}

		parser := infrastructure.NewParser(report)
		result := parser.Parse()

		// Формирование payload для API
		payload := buildSubmitPayload(submitIP, result)

		// Отправка в API
		if err := submitToAPI(payload); err != nil {
			logrus.WithError(err).Error("Failed to submit to API")
			return
		}

		// Отображение результата
		displayResult(submitIP, result)
		logrus.Info("Successfully submitted to API")
	},
}

type SubmitPayload struct {
	IP         string         `json:"ip"`
	Score      int            `json:"score"`
	SpamTrap   int            `json:"spam_trap"`
	Blocklists string         `json:"blocklists"`
	Complaints string         `json:"complaints"`
	History    []HistoryEntry `json:"history"`
}

type HistoryEntry struct {
	Date     string `json:"date"`
	Score    int    `json:"score"`
	Volume   int    `json:"volume"`
	SpamTrap int    `json:"spam_trap"`
}

func buildSubmitPayload(ip string, result *infrastructure.Result) SubmitPayload {
	volumes := make(map[string]int)
	for _, v := range result.SSVolume {
		volumes[v.Timestamp] = v.Value
	}

	history := make([]HistoryEntry, 0, len(result.SSTrend))
	for _, p := range result.SSTrend {
		ms, _ := strconv.ParseInt(p.Timestamp, 10, 64)
		tm := time.Unix(0, ms*int64(time.Millisecond))

		history = append(history, HistoryEntry{
			Date:     tm.Format("02.01.2006"),
			Score:    p.Value,
			Volume:   volumes[p.Timestamp],
			SpamTrap: result.SpamTrap,
		})
	}

	return SubmitPayload{
		IP:         ip,
		Score:      result.SenderScore,
		SpamTrap:   result.SpamTrap,
		Blocklists: result.Blocklists,
		Complaints: result.Complaints,
		History:    history,
	}
}

func submitToAPI(payload SubmitPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/scores/submit", submitAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", submitAPIToken))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Debug("API response")

	return nil
}

func displayResult(ip string, result *infrastructure.Result) {
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
}
