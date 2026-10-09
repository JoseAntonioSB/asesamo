package user_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/user"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	updatedTheme string
	updateErr    error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

func (m *mockUserRepo) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedTheme = theme
	return nil
}

func TestUpdateThemeUseCase_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uc := user.NewUpdateThemeUseCase(repo)

	userID := uuid.New()
	res, err := uc.Execute(context.Background(), user.UpdateThemeInput{
		UserID: userID,
		Theme:  "light",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "light" {
		t.Errorf("expected light, got %s", res)
	}
	if repo.updatedTheme != "light" {
		t.Errorf("expected repo theme light, got %s", repo.updatedTheme)
	}
}

func TestUpdateThemeUseCase_InvalidTheme(t *testing.T) {
	repo := &mockUserRepo{}
	uc := user.NewUpdateThemeUseCase(repo)

	_, err := uc.Execute(context.Background(), user.UpdateThemeInput{
		UserID: uuid.New(),
		Theme:  "neon-pink",
	})
	if err == nil {
		t.Fatal("expected error for invalid theme")
	}
	if !domainErrors.IsInvalidInput(err) {
		t.Errorf("expected invalid input error, got %v", err)
	}
}
