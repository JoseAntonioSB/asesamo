package user_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepoFull struct {
	mock.Mock
}

func (m *mockUserRepoFull) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepoFull) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepoFull) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

func (m *mockUserRepoFull) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	args := m.Called(ctx, id, theme)
	return args.Error(0)
}

func (m *mockUserRepoFull) UpdateProfile(ctx context.Context, id uuid.UUID, fullName string, phone *string) error {
	args := m.Called(ctx, id, fullName, phone)
	return args.Error(0)
}

func (m *mockUserRepoFull) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarKey string) error {
	args := m.Called(ctx, id, avatarKey)
	return args.Error(0)
}

func (m *mockUserRepoFull) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

// Ensure interface satisfaction
var _ repository.UserRepository = (*mockUserRepoFull)(nil)

type mockContractorRepoForProfile struct {
	mock.Mock
}

func (m *mockContractorRepoForProfile) CreateContractor(ctx context.Context, c *entity.Contractor) error {
	return nil
}
func (m *mockContractorRepoForProfile) GetContractorByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contractor), args.Error(1)
}
func (m *mockContractorRepoForProfile) GetContractorByRFC(ctx context.Context, rfc string) (*entity.Contractor, error) {
	return nil, nil
}
func (m *mockContractorRepoForProfile) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	args := m.Called(ctx, adminUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contractor), args.Error(1)
}
func (m *mockContractorRepoForProfile) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]repository.ContractorWithDetails, error) {
	return nil, nil
}
func (m *mockContractorRepoForProfile) CreatePlantContractor(ctx context.Context, pc *entity.PlantContractor) error {
	return nil
}
func (m *mockContractorRepoForProfile) UpdatePlantContractorStatus(ctx context.Context, plantID, contractorID uuid.UUID, status entity.PlantContractorStatus) error {
	return nil
}
func (m *mockContractorRepoForProfile) ListWorkersByContractor(ctx context.Context, contractorID uuid.UUID) ([]repository.ContractorWorkerDetail, error) {
	return nil, nil
}
func (m *mockContractorRepoForProfile) AddWorker(ctx context.Context, workerUser *entity.User, cw *entity.ContractorWorker) error {
	return nil
}
func (m *mockContractorRepoForProfile) ToggleWorkerStatus(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error {
	return nil
}
func (m *mockContractorRepoForProfile) CreateUserWithPassword(ctx context.Context, user *entity.User, passwordHash string) error {
	return nil
}

var _ repository.ContractorRepository = (*mockContractorRepoForProfile)(nil)

type mockStorageServiceForProfile struct {
	mock.Mock
}

func (m *mockStorageServiceForProfile) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	args := m.Called(ctx, key, r, size, contentType)
	return args.Error(0)
}
func (m *mockStorageServiceForProfile) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *mockStorageServiceForProfile) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
func (m *mockStorageServiceForProfile) GetURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

var _ repository.StorageService = (*mockStorageServiceForProfile)(nil)

func generateValidPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestGetProfileUseCase(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	contractorID := uuid.New()
	phone := "5512345678"

	userObj := &entity.User{
		ID:              userID,
		TenantID:        &tenantID,
		Email:           "admin@contractor.com",
		FullName:        "Juan Perez",
		Role:            entity.RoleContractorAdmin,
		Phone:           &phone,
		ThemePreference: "dark",
		IsActive:        true,
	}

	contractorObj := &entity.Contractor{
		ID:             contractorID,
		Name:           "Metalicas SA de CV",
		RFC:            nil,
		ContractorType: entity.ContractorTypeCompany,
		AdminUserID:    &userID,
		IsActive:       true,
	}

	uRepo := new(mockUserRepoFull)
	cRepo := new(mockContractorRepoForProfile)

	uRepo.On("GetByID", mock.Anything, userID).Return(userObj, nil)
	cRepo.On("GetContractorByAdminUserID", mock.Anything, userID).Return(contractorObj, nil)

	uc := user.NewGetProfileUseCase(uRepo, cRepo)
	detail, err := uc.Execute(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, "Juan Perez", detail.User.FullName)
	assert.Equal(t, "admin@contractor.com", detail.User.Email)
	assert.NotNil(t, detail.ContractorName)
	assert.Equal(t, "Metalicas SA de CV", *detail.ContractorName)
}

func TestUpdateProfileUseCase(t *testing.T) {
	userID := uuid.New()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("OldPass123!"), bcrypt.DefaultCost)

	userObj := &entity.User{
		ID:           userID,
		Email:        "user@test.com",
		FullName:     "Juan Perez",
		PasswordHash: string(hashed),
	}

	t.Run("Update name and phone successfully", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		uRepo.On("GetByID", mock.Anything, userID).Return(userObj, nil)
		uRepo.On("UpdateProfile", mock.Anything, userID, "Juan Carlos Perez", mock.Anything).Return(nil)

		uc := user.NewUpdateProfileUseCase(uRepo)
		phone := "5598765432"
		err := uc.Execute(context.Background(), user.UpdateProfileInput{
			UserID:   userID,
			FullName: "Juan Carlos Perez",
			Phone:    &phone,
		})

		require.NoError(t, err)
		uRepo.AssertExpectations(t)
	})

	t.Run("Update password with valid current password", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		uRepo.On("GetByID", mock.Anything, userID).Return(userObj, nil)
		uRepo.On("UpdateProfile", mock.Anything, userID, "Juan Perez", mock.Anything).Return(nil)
		uRepo.On("UpdatePassword", mock.Anything, userID, mock.Anything).Return(nil)

		uc := user.NewUpdateProfileUseCase(uRepo)
		err := uc.Execute(context.Background(), user.UpdateProfileInput{
			UserID:          userID,
			FullName:        "Juan Perez",
			CurrentPassword: "OldPass123!",
			NewPassword:     "NewPass123!Secure",
		})

		require.NoError(t, err)
		uRepo.AssertExpectations(t)
	})

	t.Run("Fail password update if current password wrong", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		uRepo.On("GetByID", mock.Anything, userID).Return(userObj, nil)

		uc := user.NewUpdateProfileUseCase(uRepo)
		err := uc.Execute(context.Background(), user.UpdateProfileInput{
			UserID:          userID,
			FullName:        "Juan Perez",
			CurrentPassword: "WrongPassword!",
			NewPassword:     "NewPass123!Secure",
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	})
}

func TestUploadAvatarUseCase(t *testing.T) {
	callerID := uuid.New()
	pngData := generateValidPNG()

	userObj := &entity.User{
		ID:       callerID,
		Email:    "juan@example.com",
		FullName: "Juan",
	}

	t.Run("Success upload own avatar", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		storage := new(mockStorageServiceForProfile)

		uRepo.On("GetByID", mock.Anything, callerID).Return(userObj, nil)
		storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, int64(len(pngData)), "image/png").Return(nil)
		uRepo.On("UpdateAvatar", mock.Anything, callerID, mock.Anything).Return(nil)

		uc := user.NewUploadAvatarUseCase(uRepo, storage)
		res, err := uc.Execute(context.Background(), user.UploadAvatarInput{
			TargetUserID: callerID,
			CallerID:     callerID,
			CallerRole:   string(entity.RoleWorker),
			FileName:     "avatar.png",
			File:         bytes.NewReader(pngData),
			FileSize:     int64(len(pngData)),
		})

		require.NoError(t, err)
		assert.NotEmpty(t, res)
		assert.Contains(t, res, "avatars/")
	})

	t.Run("Reject invalid file type (magic bytes mismatch)", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		storage := new(mockStorageServiceForProfile)

		uRepo.On("GetByID", mock.Anything, callerID).Return(userObj, nil)

		uc := user.NewUploadAvatarUseCase(uRepo, storage)
		fakeData := []byte("plain text not image")
		_, err := uc.Execute(context.Background(), user.UploadAvatarInput{
			TargetUserID: callerID,
			CallerID:     callerID,
			CallerRole:   string(entity.RoleWorker),
			FileName:     "avatar.png",
			File:         bytes.NewReader(fakeData),
			FileSize:     int64(len(fakeData)),
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidFileType)
	})

	t.Run("Forbidden when worker tries to upload for another worker", func(t *testing.T) {
		uRepo := new(mockUserRepoFull)
		storage := new(mockStorageServiceForProfile)
		otherUserID := uuid.New()

		uc := user.NewUploadAvatarUseCase(uRepo, storage)
		_, err := uc.Execute(context.Background(), user.UploadAvatarInput{
			TargetUserID: otherUserID,
			CallerID:     callerID,
			CallerRole:   string(entity.RoleWorker),
			FileName:     "avatar.png",
			File:         bytes.NewReader(pngData),
			FileSize:     int64(len(pngData)),
		})

		assert.ErrorIs(t, err, domainErrors.ErrForbidden)
	})
}
