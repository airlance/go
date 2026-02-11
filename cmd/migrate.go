package cmd

import (
	"git.emercury.dev/emercury/senderscore/api/internal/data"
	"git.emercury.dev/emercury/senderscore/api/internal/infrastructure"
	"git.emercury.dev/emercury/senderscore/api/pkg/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateCmd = cobra.Command{
	Use:   "migrate",
	Short: "Run versioned database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		db, err := infrastructure.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}

		m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
			{
				ID: "202602091900_create_base_tables",
				Migrate: func(tx *gorm.DB) error {
					if err := tx.AutoMigrate(&data.GroupModel{}); err != nil {
						return err
					}
					if err := tx.AutoMigrate(&data.IPModel{}); err != nil {
						return err
					}
					if err := tx.AutoMigrate(&data.GroupIPModel{}); err != nil {
						return err
					}
					if err := tx.AutoMigrate(&data.HistoryModel{}); err != nil {
						return err
					}
					if err := tx.AutoMigrate(&data.ScoreStatModel{}); err != nil {
						return err
					}
					return nil
				},
				Rollback: func(tx *gorm.DB) error {
					return tx.Migrator().DropTable(
						"sender_score_score_stats",
						"sender_score_histories",
						"sender_score_group_ips",
						"sender_score_ips",
						"sender_score_groups",
					)
				},
			},
		})

		if err := m.Migrate(); err != nil {
			logrus.Fatalf("Could not migrate: %v", err)
		}
		logrus.Info("Migration run successfully")
	},
}
