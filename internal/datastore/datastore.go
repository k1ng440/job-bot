package datastore

import (
	"context"
)

type JobPosting struct {
	Platform string
	ID       string
	Url      string
	Title    string
	Company  string
	Applied  bool
}

type Datastore interface {
	IncAppliedTodayCount(ctx context.Context, platform string) error
	GetAppliedTodayCount(ctx context.Context) (int, error)
	GetAppliedCountByCompany(ctx context.Context, name string) (int, error)
	IncAppliedCountByCompany(ctx context.Context, name string) error
	InsertJobPosting(ctx context.Context, jobPosting *JobPosting) error
	GetUnappliedJobPosting(ctx context.Context) (*JobPosting, error)
	Close() error
}
