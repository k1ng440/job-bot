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

package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type Config struct {
	LinkedIn struct {
		Username   string
		Password   string
		SearchUrls []string
	}
	Headless bool
}

func main() {
	dir, err := os.MkdirTemp("", "chromedp-example")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp dir")
		return
	}
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(dir),
		chromedp.NoSandbox,
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	login(taskCtx)
}

func login(ctx context.Context) {
	var title string
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.linkedin.com/login"),
		chromedp.Title(&title),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to navigate to login page")
		return
	}

	if !strings.Contains(title, "LinkedIn Login") {
		log.Info().Msg("Login is not required. Continuing...")
		return
	}

	log.Info().
		Str("title", title).
		Msg("Login required")
	if err := chromedp.Run(ctx,
		chromedp.WaitReady(`#username`, chromedp.ByID),
		chromedp.SendKeys(`#username`, "username@something.com", chromedp.ByID),
		chromedp.SendKeys(`#password`, "password", chromedp.ByID),
		chromedp.Click(`button[type="submit"]`, chromedp.NodeVisible),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to login")
		return
	}

	if err := chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second),
		chromedp.Title(&title),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to get title")
		return
	}

	if strings.Contains(title, "Security Verification") {
		log.Fatal().Msg("Security verification is required. Re-run the program in non headless and try again")
		return
	}

	if strings.Contains(title, "LinkedIn Login") {
		log.Fatal().Msg("Login failed. Please check your credentials")
		return
	}

	log.Info().Msg("Login successful")
}
