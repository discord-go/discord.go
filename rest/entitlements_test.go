package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestListSKUs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/skus" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","name":"sku1"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	skus, err := c.ListSKUs(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(skus) != 1 || skus[0].ID.String() != "456" {
		t.Errorf("Expected 1 sku with id 456, got %v", skus)
	}
}

func TestListEntitlements(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/entitlements" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","type":8}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	entitlements, err := c.ListEntitlements(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entitlements) != 1 || entitlements[0].ID.String() != "456" {
		t.Errorf("Expected 1 entitlement with id 456, got %v", entitlements)
	}
}

func TestGetEntitlement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/entitlements/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","type":8}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("123")
	entID, _ := snowflake.Parse("456")
	entitlement, err := c.GetEntitlement(context.Background(), appID, entID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if entitlement.ID.String() != "456" {
		t.Errorf("Expected entitlement id 456, got %s", entitlement.ID.String())
	}
}

func TestCreateTestEntitlement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/entitlements" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","type":8}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("123")
	skuID, _ := snowflake.Parse("789")
	ownerID, _ := snowflake.Parse("101")
	entitlement, err := c.CreateTestEntitlement(context.Background(), appID, CreateTestEntitlementParams{SKUID: skuID, OwnerID: ownerID, OwnerType: 1})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if entitlement.ID.String() != "456" {
		t.Errorf("Expected entitlement id 456, got %s", entitlement.ID.String())
	}
}

func TestDeleteTestEntitlement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/entitlements/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("123")
	entID, _ := snowflake.Parse("456")
	err := c.DeleteTestEntitlement(context.Background(), appID, entID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestConsumeEntitlement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/123/entitlements/456/consume" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("123")
	entID, _ := snowflake.Parse("456")
	err := c.ConsumeEntitlement(context.Background(), appID, entID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
