package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	pg "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/infra/postgres"
	promptusecase "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/usecase"
	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
)

const currentUserKey = "currentUser"

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-handler-create", "handler-create@example.com", "Handler Create User")
	handler := NewHandler(promptusecase.New(pg.NewRepository(db)))
	recorder := httptest.NewRecorder()

	body := bytes.NewBufferString(`{"name":"  My Prompt  ","body":"  Prompt body  "}`)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/prompts", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(currentUserKey, &userdomain.User{ID: userID})

	handler.Create(c)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var got response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.Name != "My Prompt" {
		t.Fatalf("Name = %q", got.Name)
	}
	if got.Body != "Prompt body" {
		t.Fatalf("Body = %q", got.Body)
	}
	if !got.IsActive {
		t.Fatal("expected IsActive = true")
	}
	if got.IsDefault {
		t.Fatal("expected IsDefault = false")
	}
}

func TestList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-handler-list", "handler-list@example.com", "Handler List User")
	otherUserID := testutil.SeedUser(t, db, "google", "sub-handler-list-other", "handler-list-other@example.com", "Handler List Other User")
	testutil.SeedPrompt(t, db, &userID, "my-active", "visible", true, false)
	testutil.SeedPrompt(t, db, &userID, "my-inactive", "hidden", false, false)
	testutil.SeedPrompt(t, db, &otherUserID, "other-active", "hidden", true, false)

	handler := NewHandler(promptusecase.New(pg.NewRepository(db)))
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/prompts", nil)
	c.Set(currentUserKey, &userdomain.User{ID: userID})

	handler.List(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var got []response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	var foundSystem bool
	var foundMine bool
	for _, prompt := range got {
		switch prompt.Name {
		case "ブログ記事作成（標準）":
			foundSystem = true
		case "my-active":
			foundMine = true
		case "my-inactive", "other-active":
			t.Fatalf("unexpected prompt in list: %q", prompt.Name)
		}
	}

	if !foundSystem {
		t.Fatal("expected seeded system prompt in list")
	}
	if !foundMine {
		t.Fatal("expected active user prompt in list")
	}
}

func TestUpdate_ForbiddenForSystemPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-handler-update", "handler-update@example.com", "Handler Update User")
	systemPromptID := testutil.SeedPrompt(t, db, nil, "system-readonly", "body", true, true)
	handler := NewHandler(promptusecase.New(pg.NewRepository(db)))
	recorder := httptest.NewRecorder()

	body := bytes.NewBufferString(`{"name":"updated"}`)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPatch, "/prompts/1", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set(currentUserKey, &userdomain.User{ID: userID})
	c.Params[0].Value = strconvFormatInt(systemPromptID)

	handler.Update(c)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
}

func TestDelete_ForbiddenForOtherUsersPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-handler-delete", "handler-delete@example.com", "Handler Delete User")
	otherUserID := testutil.SeedUser(t, db, "google", "sub-handler-delete-other", "handler-delete-other@example.com", "Handler Delete Other User")
	otherPromptID := testutil.SeedPrompt(t, db, &otherUserID, "other", "body", true, false)
	handler := NewHandler(promptusecase.New(pg.NewRepository(db)))
	recorder := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodDelete, "/prompts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: strconvFormatInt(otherPromptID)}}
	c.Set(currentUserKey, &userdomain.User{ID: userID})

	handler.Delete(c)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
}

func strconvFormatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}
