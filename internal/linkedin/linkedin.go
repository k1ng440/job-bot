package linkedin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	pcdp "github.com/chromedp/cdproto/cdp"
	cdp "github.com/chromedp/chromedp"
	"github.com/k1ng440/job-bot/internal/config"
	"github.com/k1ng440/job-bot/internal/datastore"
	"github.com/k1ng440/job-bot/internal/utils"
	"github.com/rs/zerolog/log"
)

type regex struct {
	title       []*regexp.Regexp
	company     []*regexp.Regexp
	description []*regexp.Regexp
}

type Linkedin struct {
	ds     datastore.Datastore
	regex  *regex
	config config.Linkedin
}

var (
	ErrBlacklisted        = errors.New("blacklisted")
	ErrSecurityCheck      = errors.New("security check")
	ErrInvalidCredentials = errors.New("invalid credentials")
	jobIdRegex            = regexp.MustCompile(`\/view\/([0-9]+)\/`)
)

const (
	platform = "linkedin"
)

func New(cfg config.Linkedin, ds datastore.Datastore) *Linkedin {
	l := &Linkedin{
		config: cfg,
		ds:     ds,
		regex: &regex{
			title:       []*regexp.Regexp{},
			company:     []*regexp.Regexp{},
			description: []*regexp.Regexp{},
		},
	}

	// Compile regex patterns
	// TODO: Handle errors here to provide better feedback

	for _, t := range cfg.Blacklists.Title {
		l.regex.title = append(l.regex.title, regexp.MustCompile(t))
	}

	for _, c := range cfg.Blacklists.Company {
		l.regex.company = append(l.regex.company, regexp.MustCompile(c))
	}

	for _, d := range cfg.Blacklists.Description {
		l.regex.description = append(l.regex.description, regexp.MustCompile(d))
	}

	return l
}

func (l *Linkedin) Run(ctx context.Context) error {
	err := l.login(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to login to linkedin")
		return err
	}

	for _, url := range l.config.SearchUrls {
		log.Info().Str("url", url).Msg("searching for jobs")
		err = l.search(ctx, url)
		if err != nil {
			return err
		}
	}

	time.Sleep(10 * time.Second)
	return nil
}

func (l *Linkedin) login(ctx context.Context) error {
	var title string
	if err := cdp.Run(ctx,
		cdp.Navigate("https://www.linkedin.com/login"),
		cdp.Title(&title),
	); err != nil {
		return fmt.Errorf("failed to retrieve to login page: %w", err)
	}

	if !strings.Contains(title, "LinkedIn Login") {
		log.Info().Msg("Login is not required. Continuing...")
		return nil
	}

	log.Info().Str("title", title).Msg("Login required")
	if err := cdp.Run(ctx,
		cdp.WaitReady(`#username`, cdp.ByID),
		cdp.SendKeys(`#username`, l.config.Username, cdp.ByID),
		cdp.SendKeys(`#password`, l.config.Password, cdp.ByID),
		cdp.Click(`button[type="submit"]`, cdp.NodeVisible),
	); err != nil {
		return fmt.Errorf("failed to login to linkedin. %w", err)
	}

	for i := 0; i < 5; i++ {
		var title string
		if err := cdp.Run(ctx,
			cdp.WaitNotVisible(`app-boot-bg-loader`, cdp.ByID),
			cdp.Title(&title),
		); err != nil {
			return fmt.Errorf("failed to get page title. %w", err)
		}

		switch {
		case strings.Contains(title, "LinkedIn Login"):
			// TODO: Get error message from page
			return errors.Join(
				ErrInvalidCredentials,
				errors.New("login failed. Please check your credentials"),
			)
		case strings.Contains(title, "Security Verification"):
			if l.config.Headless {
				return errors.Join(
					ErrSecurityCheck,
					errors.New("security verification is required. Re-run the program in non headless and try again"),
				)
			} else {
				log.Warn().Msg("Security verification is required. Please complete the verification")
				time.Sleep(10 * time.Second)
				continue
			}
		case strings.Contains(title, "Feed"):
			log.Info().Msg("Login successful")
			return nil
		}
	}

	return nil
}

// search searches for jobs on linkedin
func (l *Linkedin) search(ctx context.Context, u string) error {
	urlp, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("failed to parse url. %w", err)
	}

	start := 0

	length, err := l.visitSearchPage(ctx, urlp, start)
	if err != nil {
		return err
	}

	availableJobsCount, err := l.getAvailableJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available jobs. %w", err)
	}

	// Iterate over all jobs
	for start < availableJobsCount {
		log.Info().Int("start", start).Int("available_jobs", availableJobsCount).Msg("iterating over jobs")
		if start > 0 {
			length, err = l.visitSearchPage(ctx, urlp, start)
			if err != nil {
				return err
			}
		}
		start += length + 1

		// Iterate over all jobs on the page
		for i := 0; i < 7; i++ {
			applied, err := l.haveApplied(ctx, i+1)
			if err != nil {
				return err
			}

			if applied {
				log.Debug().Msg("Already applied for job")
				continue
			}

			if err := cdp.Run(ctx,
				cdp.Click(fmt.Sprintf(`(//a[contains(@class, 'job-card-list__title')])[%d]`, i+1)),
			); err != nil {
				return fmt.Errorf("failed to click on button. %w", err)
			}

			log.Debug().Msg("Clicked on job")

			if err := cdp.Run(ctx, cdp.WaitEnabled(`button.jobs-apply-button`, cdp.ByQuery)); err != nil {
				return fmt.Errorf("failed to wait for button. %w", err)
			}

			log.Debug().Msg("Job details page loaded")

			// Get job details
			// TODO: Save job details in a file
			post, err := l.parseJobDescription(ctx)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					continue
				}
				return err
			}
			if post == nil {
				continue
			}

			if err := l.apply(ctx, post); err != nil {
				log.Warn().Str("title", post.Title).Msg("Failed to apply for job")
			}
		}
	}

	return nil
}

func (l *Linkedin) apply(ctx context.Context, post *datastore.JobPosting) error {
	log.Info().Str("title", post.Title).Msg("Applying for job")

	if err := cdp.Run(ctx,
		cdp.Click(`button.jobs-apply-button`, cdp.ByQuery),
		cdp.Click(`button.jdfsaobs-apply-button`, cdp.ByQuery),
	); err != nil {
		return fmt.Errorf("failed to click on button. %w", err)
	}

	log.Debug().Msg("Clicked on apply button")

	return nil
}

func (l *Linkedin) visitSearchPage(ctx context.Context, u *url.URL, start int) (int, error) {
	length := 0
	container := `.jobs-search-results-list .job-card-container--clickable`

	if err := cdp.Run(ctx,
		cdp.Navigate(l.listUrl(u, start)),
		cdp.WaitVisible(container, cdp.ByQuery),
		cdp.Sleep(1*time.Second),
		cdp.Evaluate(`document.querySelectorAll('`+container+`').length`, &length),
	); err != nil {
		return 0, fmt.Errorf("failed to navigate to search page. %w", err)
	}

	return length, nil
}

func (l *Linkedin) getAvailableJobs(ctx context.Context) (int, error) {
	var availableJobStr string
	if err := cdp.Run(ctx, cdp.Text(`small.jobs-search-results-list__text`, &availableJobStr, cdp.ByQuery)); err != nil {
		return 0, fmt.Errorf("failed to get total number of jobs. %w", err)
	}

	availableJobCount, err := utils.ParseStringToInt(availableJobStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse total number of jobs. %w", err)
	}
	log.Info().Int("available_job", availableJobCount).Msg("Total number of jobs")

	return availableJobCount, nil
}

// haveApplied checks if the job has already been applied
func (l *Linkedin) haveApplied(ctx context.Context, index int) (bool, error) {
	// Ignore already applied jobs
	var applied string
	selector := fmt.Sprintf(`(//div[contains(@class, 'jobs-search-results-list')])[%d]`+
		`//span[contains(@class, 'tvm__text--neutral')]`, index)
	if err := cdp.Run(ctx,
		cdp.Text(selector, &applied, cdp.AtLeast(0)),
	); err != nil {
		if strings.Contains(err.Error(), "did not return any nodes") {
			return false, nil
		}
		return false, err
	}

	if strings.Contains(strings.ToLower(applied), "applied") {
		log.Debug().Msg("Job already applied. Skipping...")
		return true, nil
	}

	return false, nil
}

func (l *Linkedin) parseJobDescription(ctx context.Context) (*datastore.JobPosting, error) {
	var link []*pcdp.Node
	var title, company, description string

	if err := cdp.Run(ctx,
		cdp.Nodes(`(//div[contains(@class, 'jobs-unified-top-card__content--two-pane')]//a)[1]`, &link, cdp.AtLeast(0)),
		cdp.Text(`(//div[contains(@class, 'jobs-unified-top-card__content--two-pane')]//a)[1]/h2`, &title, cdp.AtLeast(0)),
		cdp.Text(`(//div[contains(@class, 'jobs-unified-top-card__content--two-pane')]//a)[2]`, &company, cdp.AtLeast(0)),
		cdp.Text(`//div[contains(@class, 'jobs-description-content__text')]/span`, &description, cdp.AtLeast(0)),
	); err != nil {
		log.Error().Err(err).Msg("failed to get job details")
		return nil, fmt.Errorf("failed to get job details. %w", err)
	}
	log.Debug().Msg("Job details fetched")

	post := &datastore.JobPosting{
		Platform: platform,
		Url:      link[0].AttributeValue("href"),
		ID:       jobIdRegex.FindStringSubmatch(link[0].AttributeValue("href"))[1],
		Company:  company,
		Title:    title,
	}

	err := l.filterPosting(ctx, post, description)
	if err != nil {
		if errors.Is(err, ErrBlacklisted) {
			return nil, nil
		}

		return nil, err
	}

	err = l.ds.InsertJobPosting(ctx, post)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert job posting")
		return nil, fmt.Errorf("failed to insert job posting. %w", err)
	}

	return post, nil
}

func (l *Linkedin) filterPosting(ctx context.Context, post *datastore.JobPosting, description string) error {
	// Check if the job title matches the regex pattern
	// If doesn't matches then check if the description contains the required languages
	titleBlacklisted := false
	for _, titleRegex := range l.regex.title {
		if titleRegex.MatchString(post.Title) {
			titleBlacklisted = true
			break
		}
	}

	descriptionBlacklisted := false
	for _, descRegex := range l.regex.description {
		if descRegex.MatchString(description) {
			descriptionBlacklisted = true
			break
		}
	}

	// detect language
	lang := utils.DetectLanguage(description)
	if lang == "" {
		log.Warn().Msg("Failed to detect language")
	}

	// Default to english
	if len(l.config.Languages) == 0 {
		l.config.Languages = []string{"english"}
	}

	// Check if the language is allowed
	allowedLang := false
	for _, allowed := range l.config.Languages {
		if lang == strings.ToLower(allowed) {
			allowedLang = true
			break
		}
	}

	if titleBlacklisted || descriptionBlacklisted || !allowedLang {
		log.Debug().Str("title", post.Title).Msg("Job blacklisted")
		return ErrBlacklisted
	}

	log.Debug().Str("title", post.Title).Msg("Job allowed")
	return nil
}

func (l *Linkedin) listUrl(u *url.URL, start int) string {
	query := u.Query()
	query.Set("start", strconv.Itoa(start))
	query.Set("f_AL", "true")
	u.RawQuery = query.Encode()
	return u.String()
}
