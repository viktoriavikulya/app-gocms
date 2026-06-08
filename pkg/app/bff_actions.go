package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/bff"
)

const cmsWorkspace = "root"
const cmsContentActionScope = "admin.content.write"

func newCMSActionExecutors(provider storage.StoreProvider) *bff.ExecutorRegistry {
	descriptors := bff.NewActionRegistry()
	descriptors.Register(
		bff.ActionDescriptor{
			ID:            "post.create",
			Label:         "Create post",
			Method:        "POST",
			Path:          bff.ActionPath("post.create"),
			Scope:         cmsContentActionScope,
			Capability:    modulecms.CapabilityContentWrite,
			RequiresToken: true,
		},
		bff.ActionDescriptor{
			ID:            "post.update",
			Label:         "Update post",
			Method:        "POST",
			Path:          bff.ActionPath("post.update"),
			Scope:         cmsContentActionScope,
			Capability:    modulecms.CapabilityContentWrite,
			RequiresToken: true,
		},
		bff.ActionDescriptor{
			ID:            "post.trash",
			Label:         "Trash post",
			Method:        "POST",
			Path:          bff.ActionPath("post.trash"),
			Scope:         cmsContentActionScope,
			Capability:    modulecms.CapabilityContentWrite,
			RequiresToken: true,
		},
	)
	executors := bff.NewExecutorRegistry(descriptors)
	executors.RegisterHandler("post.create", postCreateExecutor(provider))
	executors.RegisterHandler("post.update", postUpdateExecutor(provider))
	executors.RegisterHandler("post.trash", postTrashExecutor(provider))
	return executors
}

func parseActionRequest(r *http.Request) (bff.ActionRequest, error) {
	query := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return bff.ActionRequest{}, err
		}
		return bff.ActionRequest{Query: query, JSON: body}, nil
	}
	if err := r.ParseForm(); err != nil {
		return bff.ActionRequest{}, err
	}
	return bff.ActionRequest{Query: query, Form: r.PostForm}, nil
}

func postCreateExecutor(provider storage.StoreProvider) bff.ActionHandler {
	return func(ctx context.Context, _ bff.ActionDescriptor, req bff.ActionRequest) (bff.ActionResult, error) {
		if issues := bff.RequiredFields(req, "id"); len(issues) > 0 {
			return bff.ValidationFailure("", issues...), nil
		}
		entry, err := bindPostEntry(req)
		if err != nil {
			return bff.ValidationFailure(err.Error()), nil
		}
		entry.Kind = domaincontent.KindPost
		var created domaincontent.Entry
		runErr := withContentService(ctx, provider, func(ctx context.Context, content appcontent.Service) error {
			var createErr error
			created, createErr = content.CreateDraft(ctx, entry)
			if createErr != nil {
				return createErr
			}
			if entry.Status == domaincontent.StatusPublished {
				created, createErr = content.Publish(ctx, created.ID)
			}
			return createErr
		})
		if runErr != nil {
			return validationFromError(runErr), nil
		}
		result, err := bff.SuccessResult("created", created)
		if err != nil {
			return bff.ActionResult{}, err
		}
		result.Redirect = "/go-admin/posts/" + string(created.ID) + "/edit"
		return result, nil
	}
}

func postUpdateExecutor(provider storage.StoreProvider) bff.ActionHandler {
	return func(ctx context.Context, _ bff.ActionDescriptor, req bff.ActionRequest) (bff.ActionResult, error) {
		id := firstNonEmpty(req.Query["id"], firstFormValue(req, "id"))
		if id == "" {
			return bff.ValidationFailure("", bff.ValidationIssue{Field: "id", Code: "required", Message: "Post id is required"}), nil
		}
		patch, err := bindPostEntry(req)
		if err != nil {
			return bff.ValidationFailure(err.Error()), nil
		}
		var updated domaincontent.Entry
		runErr := withContentService(ctx, provider, func(ctx context.Context, content appcontent.Service) error {
			existing, ok, getErr := content.Get(ctx, domaincontent.ID(id))
			if getErr != nil {
				return getErr
			}
			if !ok || existing.Kind != domaincontent.KindPost {
				return errors.New("content not found")
			}
			updated = mergePostEntry(existing, patch)
			return content.Update(ctx, updated)
		})
		if runErr != nil {
			if runErr.Error() == "content not found" {
				return bff.ValidationFailure("", bff.ValidationIssue{Field: "id", Code: "not_found", Message: "Post not found"}), nil
			}
			return validationFromError(runErr), nil
		}
		result, err := bff.SuccessResult("updated", updated)
		if err != nil {
			return bff.ActionResult{}, err
		}
		result.Redirect = "/go-admin/posts/" + string(updated.ID) + "/edit"
		return result, nil
	}
}

func postTrashExecutor(provider storage.StoreProvider) bff.ActionHandler {
	return func(ctx context.Context, _ bff.ActionDescriptor, req bff.ActionRequest) (bff.ActionResult, error) {
		id := firstNonEmpty(req.Query["id"], firstFormValue(req, "id"))
		if id == "" {
			return bff.ValidationFailure("", bff.ValidationIssue{Field: "id", Code: "required", Message: "Post id is required"}), nil
		}
		var trashed domaincontent.Entry
		runErr := withContentService(ctx, provider, func(ctx context.Context, content appcontent.Service) error {
			var trashErr error
			trashed, trashErr = content.Trash(ctx, domaincontent.ID(id))
			return trashErr
		})
		if runErr != nil {
			return validationFromError(runErr), nil
		}
		result, err := bff.SuccessResult("trashed", trashed)
		if err != nil {
			return bff.ActionResult{}, err
		}
		result.Redirect = "/go-admin/posts"
		return result, nil
	}
}

func withContentService(ctx context.Context, provider storage.StoreProvider, fn func(context.Context, appcontent.Service) error) error {
	return provider.ForWorkspace(cmsWorkspace).WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		return fn(ctx, appcontent.NewService(appRepos, appRepos))
	})
}

func bindPostEntry(req bff.ActionRequest) (domaincontent.Entry, error) {
	if len(req.JSON) > 0 {
		var entry domaincontent.Entry
		if err := json.Unmarshal(req.JSON, &entry); err != nil {
			return domaincontent.Entry{}, err
		}
		return entry, nil
	}
	entry := domaincontent.Entry{
		Title: map[string]string{},
	}
	if value := firstFormValue(req, "id"); value != "" {
		entry.ID = domaincontent.ID(value)
	}
	if value := firstFormValue(req, "title"); value != "" {
		if strings.HasPrefix(strings.TrimSpace(value), "{") {
			if err := json.Unmarshal([]byte(value), &entry.Title); err != nil {
				return domaincontent.Entry{}, err
			}
		} else {
			entry.Title["en"] = value
		}
	}
	if value := firstFormValue(req, "slug"); value != "" {
		entry.Slug = value
	}
	if value := firstFormValue(req, "content"); value != "" {
		entry.Content = value
	}
	if value := firstFormValue(req, "excerpt"); value != "" {
		entry.Excerpt = value
	}
	if value := firstFormValue(req, "status"); value != "" {
		entry.Status = domaincontent.Status(value)
	}
	if value := firstFormValue(req, "visibility"); value != "" {
		entry.Visibility = domaincontent.Visibility(value)
	}
	if value := firstFormValue(req, "author_id"); value != "" {
		entry.AuthorID = value
	}
	if value := firstFormValue(req, "featured_media_id"); value != "" {
		entry.FeaturedMediaID = value
	}
	return entry, nil
}

func mergePostEntry(existing domaincontent.Entry, patch domaincontent.Entry) domaincontent.Entry {
	if patch.Title != nil {
		existing.Title = patch.Title
	}
	if patch.Slug != "" {
		existing.Slug = patch.Slug
	}
	if patch.Content != "" {
		existing.Content = patch.Content
	}
	if patch.Excerpt != "" {
		existing.Excerpt = patch.Excerpt
	}
	if patch.Status != "" {
		existing.Status = patch.Status
	}
	if patch.Visibility != "" {
		existing.Visibility = patch.Visibility
	}
	if patch.AuthorID != "" {
		existing.AuthorID = patch.AuthorID
	}
	if patch.FeaturedMediaID != "" {
		existing.FeaturedMediaID = patch.FeaturedMediaID
	}
	if patch.Metadata != nil {
		existing.Metadata = patch.Metadata
	}
	return existing
}

func firstFormValue(req bff.ActionRequest, key string) string {
	if req.Form == nil {
		return ""
	}
	values, ok := req.Form[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}

func validationFromError(err error) bff.ActionResult {
	return bff.ValidationFailure("", bff.ValidationIssue{Code: "invalid", Message: err.Error()})
}
