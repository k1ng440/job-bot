/* MIT License

Copyright (c) 2023 Asaduzzaman Pavel (contact@iampavel.dev)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
