package cmd

import (
	"github.com/chromedp/chromedp"
	"github.com/k1ng440/job-bot/internal/config"
	"github.com/k1ng440/job-bot/internal/datastore"
	"github.com/k1ng440/job-bot/internal/linkedin"
	"github.com/k1ng440/job-bot/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the job bot",
	Long:  ``,
	RunE:  start,
}

func start(cmd *cobra.Command, _ []string) error {
	var cfg config.Config
	viper.UnmarshalExact(&cfg)

	dir, destory := utils.Mkdir(cfg.ChromeProfilePath)
	defer destory()

	ds, err := datastore.NewSqliteDatastore("")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create datastore")
		return err
	}

	l := linkedin.New(cfg.Linkedin, ds)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(dir),
		chromedp.NoSandbox,
		chromedp.Flag("headless", cfg.Linkedin.Headless),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(cmd.Context(), opts...)
	defer cancel()

	chromedpCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := l.Run(chromedpCtx); err != nil {
		log.Error().Err(err).Msg("Linkedin bot exited with error")
		return err
	}

	return nil
}
