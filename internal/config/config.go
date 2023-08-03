package config

type Config struct {
	ChromeProfilePath string   `json:"chrome_profile_path" mapstructure:"chrome_profile_path"`
	Linkedin          Linkedin `json:"linkedin"            mapstructure:"linkedin"`
}

type Linkedin struct {
	// Username for linkedin
	Username string `json:"username" mapstructure:"username"`
	// Password for linkedin
	Password string `json:"password" mapstructure:"password"`

	// Languages is a list of languages to filter jobs by
	Languages []string `json:"languages" mapstructure:"languages"`

	// Whitelists are lists of regex patterns to match
	// If any of the patterns match, the job will be applied to

	// Blacklists are lists of regex patterns to ignore
	// If any of the patterns match, the job will be ignored
	// If the list is empty, no jobs will be ignored
	Blacklists struct {
		Title       []string `json:"title" mapstructure:"title"`             // List of regex pattern match title to ignore
		Company     []string `json:"company" mapstructure:"company"`         // List of regex pattern match company to ignore
		Description []string `json:"description" mapstructure:"description"` // List of regex pattern match job description to ignore.
	} `json:"blacklists" mapstructure:"blacklists"`

	// SearchUrls is a list of urls to search for jobs
	// Must be filtered to only show easy apply jobs
	SearchUrls []string `json:"search_urls" mapstructure:"search_urls"`

	// MaxAgeDays is the maximum age of a job posting in days
	MaxAgeDays int `json:"max_age_days" mapstructure:"max_age_days"`

	// MaxApplications is the maximum number of applications to send per day
	// This is to prevent spamming linkedin with applications
	MaxApplications int `json:"max_applications" mapstructure:"max_applications"`

	// MaxApplicationsPerCompany is the maximum number of applications to send per company per day
	// This is to prevent spamming the same company with applications
	MaxApplicationsPerCompany int `json:"max_applications_per_company" mapstructure:"max_applications_per_company"`

	// Headless is a flag to run the browser in headless mode
	Headless bool `json:"headless" mapstructure:"headless"`
}
