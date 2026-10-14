package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchServerSorting(t *testing.T) {
	tests := []struct {
		name       string
		orderField string
		orderBy    int
	}{
		{"Age ascending", "Age", OrderByAsc},
		{"Age descending", "Age", OrderByDesc},
		{"Id descending", "Id", OrderByDesc},
		{"Name ascending", "Name", OrderByAsc},
		{"Name descending", "Name", OrderByDesc},
		{"Default name", "", OrderByAsc},
		{"No sorting", "Age", OrderByAsIs},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := SearchClient{
				AccessToken: "secret",
				URL:         server.URL,
			}

			req := SearchRequest{
				Limit:      10,
				Offset:     0,
				Query:      "",
				OrderField: tt.orderField,
				OrderBy:    tt.orderBy,
			}

			result, err := client.FindUsers(req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Users) != 10 {
				t.Fatalf("expected 10 users, got %d", len(result.Users))
			}

			if tt.orderField == "Age" && tt.orderBy == OrderByAsc {
				for i := 1; i < len(result.Users); i++ {
					if result.Users[i-1].Age > result.Users[i].Age {
						t.Errorf("Ages are not sorted ascending")
					}
				}
			}

			if tt.orderField == "Age" && tt.orderBy == OrderByDesc {
				for i := 1; i < len(result.Users); i++ {
					if result.Users[i-1].Age < result.Users[i].Age {
						t.Errorf("Ages are not sorted descending")
					}
				}
			}

			if tt.orderField == "Id" && tt.orderBy == OrderByDesc {
				for i := 1; i < len(result.Users); i++ {
					if result.Users[i-1].ID < result.Users[i].ID {
						t.Errorf("IDs are not sorted descending")
					}
				}
			}

			if (tt.orderField == "Name" || tt.orderField == "") && tt.orderBy == OrderByAsc {
				for i := 1; i < len(result.Users); i++ {
					if result.Users[i-1].Name > result.Users[i].Name {
						t.Errorf("names are not sorted ascending")
					}
				}
			}

			if tt.orderBy == OrderByAsIs {
				for i, user := range result.Users {
					if user.ID != i {
						t.Errorf("expected ID %d, got %d", i, user.ID)
					}
				}
			}

		})
	}
}

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

func TestSearchServerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	tests := []struct {
		name        string
		token       string
		orderField  string
		expectedErr string
	}{
		{
			name:        "Invalid token",
			token:       "wrong",
			orderField:  "Id",
			expectedErr: "bad AccessToken",
		},
		{
			name:        "Invalid order field",
			token:       "secret",
			orderField:  "Salary",
			expectedErr: "OrderFeld Salary invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := SearchClient{
				AccessToken: tt.token,
				URL:         server.URL,
			}

			req := SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "",
				OrderField: tt.orderField,
				OrderBy:    OrderByAsc,
			}

			_, err := client.FindUsers(req)

			if err == nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}

		})
	}
}

func TestSearchClientInvalidParams(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		offset      int
		expectedErr string
	}{
		{
			name:        "Negative limit",
			limit:       -1,
			offset:      0,
			expectedErr: "limit must be > 0",
		},
		{
			name:        "Negative offset",
			limit:       5,
			offset:      -1,
			expectedErr: "offset must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := SearchClient{}

			req := SearchRequest{
				Limit:      tt.limit,
				Offset:     tt.offset,
				Query:      "",
				OrderField: "Id",
				OrderBy:    OrderByAsc,
			}

			_, err := client.FindUsers(req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}

}

func TestSearchServerFileError(t *testing.T) {
	oldPath := datasetPath
	datasetPath = "nonexistent_file.xml"

	defer func() {
		datasetPath = oldPath
	}()

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

	_, err := client.FindUsers(req)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := "SearchServer fatal error"

	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestSearchServerInvalidXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.xml")

	err := os.WriteFile(path, []byte("<root><row>"), 0644)
	if err != nil {
		t.Fatalf("failed to create XML file: %v", err)
	}

	oldPath := datasetPath
	datasetPath = path
	defer func() {
		datasetPath = oldPath
	}()

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

	_, err = client.FindUsers(req)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := "SearchServer fatal error"

	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestSearchServerPagination(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "Offset beyond users",
			limit:         5,
			offset:        10000,
			expectedCount: 0,
		},
		{
			name:          "Limit beyond remaining users",
			limit:         10,
			offset:        34,
			expectedCount: 1,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := SearchClient{
				AccessToken: "secret",
				URL:         server.URL,
			}

			req := SearchRequest{
				Limit:      tt.limit,
				Offset:     tt.offset,
				Query:      "",
				OrderField: "Id",
				OrderBy:    OrderByAsc,
			}

			result, err := client.FindUsers(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Users) != tt.expectedCount {
				t.Errorf("expected %d users, got %d", tt.expectedCount, len(result.Users))
			}

			if tt.name == "Offset beyond users" && result.NextPage {
				t.Errorf("expected NextPage=false, got true")
			}

		})
	}

}

func TestSearchServerInvalidPagination(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{"Negative limit", -1, 0},
		{"Negative offset", 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf(
				"/?limit=%d&offset=%d&order_field=Id&order_by=1",
				tt.limit,
				tt.offset,
			)

			req := httptest.NewRequest(http.MethodGet, url, nil)

			req.Header.Set("AccessToken", "secret")

			w := httptest.NewRecorder()

			SearchServer(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestSearchServerNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		AccessToken: "secret",
		URL:         server.URL,
	}

	req := SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "xyz123nonexistent",
		OrderField: "Name",
		OrderBy:    OrderByAsc,
	}

	result, err := client.FindUsers(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Users) != 0 {
		t.Errorf("expected %d users, got %d", 0, len(result.Users))
	}

	if result.NextPage {
		t.Errorf("expected NextPage=false, got true")
	}

}

func TestSearchServerInvalidNumbers(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "Invalid order_by",
			url:  "/?limit=5&offset=0&order_field=Id&order_by=abc",
		},
		{
			name: "Invalid limit",
			url:  "/?limit=abc&offset=0&order_field=Id&order_by=1",
		},
		{
			name: "Invalid offset",
			url:  "/?limit=5&offset=abc&order_field=Id&order_by=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("AccessToken", "secret")

			w := httptest.NewRecorder()
			SearchServer(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestSearchClientNextPage(t *testing.T) {
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

	if !result.NextPage {
		t.Error("expected NextPage=true, got false")
	}
}
