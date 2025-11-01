package happyeyeballs

import (
	"context"
	"fmt"
	"os"
)

type SupabaseStore struct {
	url    string
	apiKey string
	client interface{}
}

func NewSupabaseStore() (*SupabaseStore, error) {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_ANON_KEY")

	if url == "" || apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY environment variables must be set")
	}

	return &SupabaseStore{
		url:    url,
		apiKey: apiKey,
	}, nil
}

func (s *SupabaseStore) SaveConnectionAttempt(ctx context.Context, attempt *ConnectionAttempt) error {
	return nil
}

func (s *SupabaseStore) SaveDNSResolution(ctx context.Context, resolution *DNSResolution) error {
	return nil
}
