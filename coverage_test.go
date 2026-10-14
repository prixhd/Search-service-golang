package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchServerSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		AccessToken: "secret",
		URL:         server.URL,
	}

	req := SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	result, err := client.FindUsers(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Users) != 5 {
		t.Errorf("expected 5 users, got %d", len(result.Users))
	}
}
