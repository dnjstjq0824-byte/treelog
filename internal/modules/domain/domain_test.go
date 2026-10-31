package domain

import (
	"context"
	"testing"
)

func TestDomainValidation(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name    string
		input   RegisterDomainInput
		wantErr bool
	}{
		{
			name:    "empty domain",
			input:   RegisterDomainInput{DomainName: ""},
			wantErr: true,
		},
		{
			name:    "no dot in domain",
			input:   RegisterDomainInput{DomainName: "invalid-domain"},
			wantErr: true,
		},
		{
			name:    "domain with space",
			input:   RegisterDomainInput{DomainName: "invalid domain.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Without DB pool, validation failure returns Fault before DB call
			_, err := svc.RegisterDomain(context.Background(), "blog-id", "user-id", tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}
