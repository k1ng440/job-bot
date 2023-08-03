package datastore_test

import (
	"context"
	"os"
	"testing"

	"github.com/k1ng440/job-bot/internal/datastore"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
)

// testDBFile is the temporary SQLite database file used for testing.
const testDBFile = "test.db"

// setupDB initializes a test database and returns a datastore.Datastore for testing.
func setupDB(t *testing.T) (datastore.Datastore, func()) {
	// Open the database
	ds, err := datastore.NewSqliteDatastore(testDBFile)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	cleanup := func() {
		ds.Close()
		os.Remove(testDBFile)
	}

	return ds, cleanup
}

func TestInsertJobPosting(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Create a test JobPosting
	jobPosting := &datastore.JobPosting{
		Platform: "TestPlatform",
		ID:       "123",
		Url:      "https://test.com/job/testid",
		Title:    "Test Job",
		Company:  "Test Company",
		Applied:  false,
	}

	// Insert the test JobPosting
	err := ds.InsertJobPosting(context.Background(), jobPosting)
	if err != nil {
		t.Fatalf("failed to insert job posting: %v", err)
	}

	// Retrieve the inserted JobPosting from the database
	retrievedJobPosting, err := ds.GetUnappliedJobPosting(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve job posting: %v", err)
	}

	// Check that the retrieved JobPosting matches the inserted one
	if retrievedJobPosting == nil {
		t.Fatal("no job posting found in the database")
	}
	if retrievedJobPosting.Platform != jobPosting.Platform ||
		retrievedJobPosting.ID != jobPosting.ID ||
		retrievedJobPosting.Url != jobPosting.Url ||
		retrievedJobPosting.Title != jobPosting.Title ||
		retrievedJobPosting.Company != jobPosting.Company ||
		retrievedJobPosting.Applied != jobPosting.Applied {
		t.Fatal("retrieved job posting does not match the inserted one")
	}
}

func TestIncAppliedTodayCount(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Increment applied count for a platform
	platform := "TestPlatform"
	err := ds.IncAppliedTodayCount(context.Background(), platform)
	if err != nil {
		t.Fatalf("failed to increment applied count: %v", err)
	}

	// Retrieve the applied count for the platform
	count, err := ds.GetAppliedTodayCount(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve applied count: %v", err)
	}

	// Check that the retrieved count is 1
	if count != 1 {
		t.Fatalf("expected applied count to be 1, got %d", count)
	}
}

func TestIncAppliedCountByCompany(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Increment applied count for a company
	company := "TestCompany"
	err := ds.IncAppliedCountByCompany(context.Background(), company)
	if err != nil {
		t.Fatalf("failed to increment applied count by company: %v", err)
	}

	// Retrieve the applied count for the company
	count, err := ds.GetAppliedCountByCompany(context.Background(), company)
	if err != nil {
		t.Fatalf("failed to retrieve applied count by company: %v", err)
	}

	// Check that the retrieved count is 1
	if count != 1 {
		t.Fatalf("expected applied count to be 1, got %d", count)
	}
}

func TestGetUnappliedJobPosting(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Insert a test JobPosting
	jobPosting := &datastore.JobPosting{
		Platform: "TestPlatform",
		ID:       "123",
		Url:      "https://example.com",
		Title:    "Test Job",
		Company:  "Test Company",
		Applied:  false,
	}
	err := ds.InsertJobPosting(context.Background(), jobPosting)
	if err != nil {
		t.Fatalf("failed to insert job posting: %v", err)
	}

	// Retrieve an unapplied job posting from the database
	retrievedJobPosting, err := ds.GetUnappliedJobPosting(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve unapplied job posting: %v", err)
	}

	// Check that the retrieved JobPosting matches the inserted one
	if retrievedJobPosting == nil {
		t.Fatal("no unapplied job posting found in the database")
	}
	if retrievedJobPosting.ID != jobPosting.ID {
		t.Fatal("retrieved job posting does not match the inserted one")
	}
}

func TestGetAppliedCountByCompany(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Increment applied count for a company
	company := "TestCompany"
	err := ds.IncAppliedCountByCompany(context.Background(), company)
	if err != nil {
		t.Fatalf("failed to increment applied count by company: %v", err)
	}

	// Retrieve the applied count for the company
	count, err := ds.GetAppliedCountByCompany(context.Background(), company)
	if err != nil {
		t.Fatalf("failed to retrieve applied count by company: %v", err)
	}

	// Check that the retrieved count is 1
	if count != 1 {
		t.Fatalf("expected applied count to be 1, got %d", count)
	}
}

func TestGetAppliedTodayCount(t *testing.T) {
	ds, cleanup := setupDB(t)
	defer cleanup()

	// Increment applied count for a platform
	platform := "TestPlatform"
	err := ds.IncAppliedTodayCount(context.Background(), platform)
	if err != nil {
		t.Fatalf("failed to increment applied count: %v", err)
	}

	// Retrieve the applied count for today
	count, err := ds.GetAppliedTodayCount(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve applied count for today: %v", err)
	}

	// Check that the retrieved count is 1
	if count != 1 {
		t.Fatalf("expected applied count to be 1, got %d", count)
	}
}
