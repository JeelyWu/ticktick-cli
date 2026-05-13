package ticktick

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProjectsReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.ListProjects(context.Background(), "token")
	if err == nil {
		t.Fatal("ListProjects() error = nil, want error")
	}
}

func TestListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := []projectDTO{
			{ID: "p1", Name: "Alpha", Kind: "TASK"},
			{ID: "p2", Name: "Beta", Kind: "NOTE"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	projects, err := client.ListProjects(context.Background(), "token")
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("len(projects) = %d, want 2", len(projects))
	}
	if projects[0].Name != "Alpha" {
		t.Fatalf("Name = %q, want Alpha", projects[0].Name)
	}
}

func TestCreateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := projectDTO{ID: "p3", Name: "Gamma", Kind: "TASK"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	project, err := client.CreateProject(context.Background(), "token", "Gamma", "TASK")
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if project.ID != "p3" {
		t.Fatalf("ID = %q, want p3", project.ID)
	}
	if project.Kind != "TASK" {
		t.Fatalf("Kind = %q, want TASK", project.Kind)
	}
}

func TestCreateProjectWithoutKind(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode error = %v", err)
		}
		if _, ok := body["kind"]; ok {
			t.Fatal("body contains kind, want omitted")
		}
		resp := projectDTO{ID: "p4", Name: "Delta"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.CreateProject(context.Background(), "token", "Delta", "")
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project/p1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		resp := projectDTO{ID: "p1", Name: "Alpha Updated"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	project, err := client.UpdateProject(context.Background(), "token", "p1", "Alpha Updated", "")
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}
	if project.Name != "Alpha Updated" {
		t.Fatalf("Name = %q, want Alpha Updated", project.Name)
	}
}

func TestDeleteProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodDelete; got != want {
			t.Fatalf("Method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/open/v1/project/p1"; got != want {
			t.Fatalf("Path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	if err := client.DeleteProject(context.Background(), "token", "p1"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
}
