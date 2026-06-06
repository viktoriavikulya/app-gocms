package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/fastygo/app-gocms/internal/appschema"
	"github.com/fastygo/platform/pkg/conformance/bffparity"
	"github.com/fastygo/platform/pkg/render"
)

func TestBFFProofScreensMatchGoldenFixtures(t *testing.T) {
	handler := testHandler(t)
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	cookie := loginCookie(t, handler, "admin", "admin")
	cases := []struct {
		path    string
		bffPath string
		fixture string
	}{
		{
			path:    "/go-admin/posts",
			bffPath: "/bff/screens/go-admin/posts",
			fixture: "appcms-posts-table.json",
		},
		{
			path:    "/go-admin/posts/post-rest/edit",
			bffPath: "/bff/screens/go-admin/posts/post-rest/edit",
			fixture: "appcms-post-edit-form.json",
		},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			inProcess, err := registry.Screen(tc.path)
			if err != nil {
				t.Fatalf("resolve screen: %v", err)
			}
			bffparity.AssertScreenMatchesFixture(t, tc.fixture, inProcess)

			response := httptest.NewRecorder()
			request := testRequest(http.MethodGet, tc.bffPath)
			request.AddCookie(cookie)
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("%s returned %d: %s", tc.bffPath, response.Code, response.Body.String())
			}
			var httpScreen render.ScreenModel
			if err := json.Unmarshal(response.Body.Bytes(), &httpScreen); err != nil {
				t.Fatalf("decode bff screen: %v", err)
			}
			bffparity.AssertScreenMatchesFixture(t, tc.fixture, httpScreen)
			if !reflect.DeepEqual(bffparity.NormalizeScreen(inProcess), bffparity.NormalizeScreen(httpScreen)) {
				t.Fatalf("in-process and HTTP screens differ after normalization:\n in-process: %#v\n http: %#v", inProcess, httpScreen)
			}
			if tc.fixture == "appcms-post-edit-form.json" && httpScreen.Metadata["action_token"] == "" {
				t.Fatal("expected volatile action_token on HTTP form screen")
			}
		})
	}
}
