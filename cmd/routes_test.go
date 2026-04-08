package cmd

import (
	"testing"
)

func TestBuildRoutes(t *testing.T) {
	tests := []struct {
		name         string
		ports        []int
		project      string
		branch       string
		domain       string
		routerPort   int
		svcRunning   bool
		pfConfigured bool
		wantURLs     []string
	}{
		{
			name:         "router with port forwarding",
			ports:        []int{3010, 3011},
			project:      "myapp",
			branch:       "feature-x",
			domain:       "prt.dev",
			routerPort:   3001,
			svcRunning:   true,
			pfConfigured: true,
			wantURLs: []string{
				"https://myapp-feature-x.prt.dev",
				"https://myapp-feature-x.prt.dev",
			},
		},
		{
			name:         "router without port forwarding",
			ports:        []int{3010, 3011},
			project:      "myapp",
			branch:       "feature-x",
			domain:       "prt.dev",
			routerPort:   3001,
			svcRunning:   true,
			pfConfigured: false,
			wantURLs: []string{
				"https://myapp-feature-x.prt.dev:3001",
				"https://myapp-feature-x.prt.dev:3001",
			},
		},
		{
			name:         "router not running falls back to localhost",
			ports:        []int{3010, 3011},
			project:      "myapp",
			branch:       "feature-x",
			domain:       "prt.dev",
			routerPort:   3001,
			svcRunning:   false,
			pfConfigured: false,
			wantURLs: []string{
				"http://localhost:3010",
				"http://localhost:3011",
			},
		},
		{
			name:         "single port",
			ports:        []int{3010},
			project:      "api",
			branch:       "main",
			domain:       "prt.dev",
			routerPort:   3001,
			svcRunning:   true,
			pfConfigured: true,
			wantURLs: []string{
				"https://api-main.prt.dev",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes := buildRoutes(tt.ports, tt.project, tt.branch, tt.domain, tt.routerPort, tt.svcRunning, tt.pfConfigured)
			if len(routes) != len(tt.wantURLs) {
				t.Fatalf("got %d routes, want %d", len(routes), len(tt.wantURLs))
			}
			for i, r := range routes {
				if r.URL != tt.wantURLs[i] {
					t.Errorf("route[%d].URL = %q, want %q", i, r.URL, tt.wantURLs[i])
				}
				if r.Port != tt.ports[i] {
					t.Errorf("route[%d].Port = %d, want %d", i, r.Port, tt.ports[i])
				}
			}
		})
	}
}
