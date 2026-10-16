package contractor_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/contractor"

	"github.com/google/uuid"
)

// Mock ContractorRepository
type mockContractorRepo struct {
	contractors      map[uuid.UUID]*entity.Contractor
	contractorsByRFC map[string]*entity.Contractor
	plantContractors map[string]*entity.PlantContractor // key: plantID:contractorID
	workers          map[uuid.UUID]*entity.ContractorWorker
	users            map[uuid.UUID]*entity.User
}

func newMockContractorRepo() *mockContractorRepo {
	return &mockContractorRepo{
		contractors:      make(map[uuid.UUID]*entity.Contractor),
		contractorsByRFC: make(map[string]*entity.Contractor),
		plantContractors: make(map[string]*entity.PlantContractor),
		workers:          make(map[uuid.UUID]*entity.ContractorWorker),
		users:            make(map[uuid.UUID]*entity.User),
	}
}

func (m *mockContractorRepo) CreateContractor(ctx context.Context, c *entity.Contractor) error {
	m.contractors[c.ID] = c
	if c.RFC != nil {
		m.contractorsByRFC[*c.RFC] = c
	}
	return nil
}

func (m *mockContractorRepo) GetContractorByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error) {
	c, ok := m.contractors[id]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (m *mockContractorRepo) GetContractorByRFC(ctx context.Context, rfc string) (*entity.Contractor, error) {
	c, ok := m.contractorsByRFC[rfc]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (m *mockContractorRepo) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	for _, c := range m.contractors {
		if c.AdminUserID != nil && *c.AdminUserID == adminUserID {
			return c, nil
		}
	}
	return nil, domainErrors.ErrContractorNotFound
}

func (m *mockContractorRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]repository.ContractorWithDetails, error) {
	var list []repository.ContractorWithDetails
	for _, pc := range m.plantContractors {
		if pc.PlantID == plantID {
			c := m.contractors[pc.ContractorID]
			if c != nil {
				detail := repository.ContractorWithDetails{
					ID:             c.ID,
					Name:           c.Name,
					RFC:            c.RFC,
					REPSENumber:    c.REPSENumber,
					ContractorType: c.ContractorType,
					PlantStatus:    pc.Status,
					IsActive:       c.IsActive,
					CreatedAt:      c.CreatedAt,
				}
				list = append(list, detail)
			}
		}
	}
	return list, nil
}

func (m *mockContractorRepo) CreatePlantContractor(ctx context.Context, pc *entity.PlantContractor) error {
	key := pc.PlantID.String() + ":" + pc.ContractorID.String()
	m.plantContractors[key] = pc
	return nil
}

func (m *mockContractorRepo) UpdatePlantContractorStatus(ctx context.Context, plantID, contractorID uuid.UUID, status entity.PlantContractorStatus) error {
	key := plantID.String() + ":" + contractorID.String()
	pc, ok := m.plantContractors[key]
	if !ok {
		return domainErrors.ErrContractorNotFound
	}
	pc.Status = status
	return nil
}

func (m *mockContractorRepo) ListWorkersByContractor(ctx context.Context, contractorID uuid.UUID) ([]repository.ContractorWorkerDetail, error) {
	var list []repository.ContractorWorkerDetail
	for _, w := range m.workers {
		if w.ContractorID == contractorID {
			u := m.users[w.UserID]
			fullName := "Unknown"
			email := ""
			if u != nil {
				fullName = u.FullName
				email = u.Email
			}
			list = append(list, repository.ContractorWorkerDetail{
				WorkerID:       w.UserID,
				ContractorID:   w.ContractorID,
				FullName:       fullName,
				Email:          email,
				JobTitle:       w.JobTitle,
				CURP:           w.CURP,
				IMSSNumber:     w.IMSSNumber,
				IsActive:       w.IsActive,
				UserIsActive:   true,
				DocumentsCount: 0,
			})
		}
	}
	return list, nil
}

func (m *mockContractorRepo) AddWorker(ctx context.Context, workerUser *entity.User, cw *entity.ContractorWorker) error {
	m.users[workerUser.ID] = workerUser
	m.workers[cw.ID] = cw
	return nil
}

func (m *mockContractorRepo) ToggleWorkerStatus(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error {
	for _, w := range m.workers {
		if w.ContractorID == contractorID && w.UserID == workerUserID {
			w.IsActive = isActive
			return nil
		}
	}
	return domainErrors.ErrWorkerNotFound
}

func (m *mockContractorRepo) CreateUserWithPassword(ctx context.Context, user *entity.User, passwordHash string) error {
	m.users[user.ID] = user
	return nil
}

// Mock UserRepository
type mockUserRepo struct {
	users map[string]*entity.User // email -> user
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*entity.User)}
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domainErrors.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, domainErrors.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

func (m *mockUserRepo) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	return nil
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, fullName string, phone *string) error {
	return nil
}

func (m *mockUserRepo) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarKey string) error {
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

// Mock PlantRepository
type mockPlantRepo struct {
	plants map[uuid.UUID]*entity.Plant
}

func newMockPlantRepo() *mockPlantRepo {
	return &mockPlantRepo{plants: make(map[uuid.UUID]*entity.Plant)}
}

func (m *mockPlantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	p, ok := m.plants[id]
	if !ok {
		return nil, domainErrors.ErrPlantNotFound
	}
	return p, nil
}

func (m *mockPlantRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error) {
	var list []entity.Plant
	for _, p := range m.plants {
		if p.TenantID == tenantID {
			list = append(list, *p)
		}
	}
	return list, nil
}

// Tests for InviteContractorUseCase
func TestInviteContractorUseCase(t *testing.T) {
	ctx := context.Background()
	cRepo := newMockContractorRepo()
	uRepo := newMockUserRepo()
	pRepo := newMockPlantRepo()

	plantID := uuid.New()
	tenantID := uuid.New()
	pRepo.plants[plantID] = &entity.Plant{
		ID:       plantID,
		TenantID: tenantID,
		Name:     "Planta Monterrey",
	}

	uc := contractor.NewInviteContractorUseCase(cRepo, uRepo, pRepo)

	t.Run("fails when name is missing", func(t *testing.T) {
		_, err := uc.Execute(ctx, contractor.InviteContractorInput{
			PlantID:    plantID,
			RFC:        "ABC123456XYZ",
			AdminEmail: "admin@corp.mx",
		})
		if err == nil {
			t.Fatal("expected error for missing name")
		}
	})

	t.Run("fails when RFC is missing", func(t *testing.T) {
		_, err := uc.Execute(ctx, contractor.InviteContractorInput{
			Name:       "Acme Corp",
			PlantID:    plantID,
			AdminEmail: "admin@corp.mx",
		})
		if err == nil {
			t.Fatal("expected error for missing RFC")
		}
	})

	t.Run("fails when AdminEmail is missing", func(t *testing.T) {
		_, err := uc.Execute(ctx, contractor.InviteContractorInput{
			Name:    "Acme Corp",
			RFC:     "ABC123456XYZ",
			PlantID: plantID,
		})
		if err == nil {
			t.Fatal("expected error for missing AdminEmail")
		}
	})

	t.Run("fails when Plant does not exist", func(t *testing.T) {
		_, err := uc.Execute(ctx, contractor.InviteContractorInput{
			Name:       "Acme Corp",
			RFC:        "ABC123456XYZ",
			AdminEmail: "admin@corp.mx",
			PlantID:    uuid.New(),
		})
		if err == nil {
			t.Fatal("expected error for non-existent plant")
		}
	})

	t.Run("successfully invites new contractor and creates admin", func(t *testing.T) {
		repse := "REPSE-2024-998877"
		res, err := uc.Execute(ctx, contractor.InviteContractorInput{
			TenantID:       tenantID,
			PlantID:        plantID,
			Name:           "Seguridad Industrial del Norte SA de CV",
			RFC:            "SIN123456AB1",
			REPSENumber:    repse,
			ContractorType: "company",
			AdminEmail:     "contacto@seguridadnorte.mx",
			AdminFullName:  "Carlos Ramirez",
			AdminPhone:     "8181234567",
			InitialStatus:  "active",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatal("expected non-nil contractor")
		}
		if res.Name != "Seguridad Industrial del Norte SA de CV" {
			t.Errorf("got name %s", res.Name)
		}
		if *res.RFC != "SIN123456AB1" {
			t.Errorf("got RFC %s", *res.RFC)
		}
		if res.AdminUserID == nil {
			t.Fatal("expected admin user id to be set")
		}
	})
}

// Tests for ListPlantContractorsUseCase and UpdateContractorStatusUseCase
func TestListAndUpdatePlantContractors(t *testing.T) {
	ctx := context.Background()
	cRepo := newMockContractorRepo()
	pRepo := newMockPlantRepo()

	plantID := uuid.New()
	contractorID := uuid.New()
	cRepo.contractors[contractorID] = &entity.Contractor{
		ID:       contractorID,
		Name:     "Mantenimiento Total",
		IsActive: true,
	}
	cRepo.plantContractors[plantID.String()+":"+contractorID.String()] = &entity.PlantContractor{
		PlantID:      plantID,
		ContractorID: contractorID,
		Status:       entity.PlantContractorActive,
	}

	listUC := contractor.NewListPlantContractorsUseCase(cRepo, pRepo)
	updateUC := contractor.NewUpdateContractorStatusUseCase(cRepo)

	t.Run("lists contractors for plant", func(t *testing.T) {
		list, err := listUC.Execute(ctx, plantID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 contractor, got %d", len(list))
		}
		if list[0].Name != "Mantenimiento Total" {
			t.Errorf("got contractor %s", list[0].Name)
		}
	})

	t.Run("updates contractor status", func(t *testing.T) {
		err := updateUC.Execute(ctx, plantID, contractorID, "suspended")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		key := plantID.String() + ":" + contractorID.String()
		if cRepo.plantContractors[key].Status != entity.PlantContractorSuspended {
			t.Fatalf("expected suspended, got %s", cRepo.plantContractors[key].Status)
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		err := updateUC.Execute(ctx, plantID, contractorID, "invalid_status")
		if err == nil {
			t.Fatal("expected error for invalid status")
		}
	})
}

// Tests for Crew Use Cases
func TestCrewUseCases(t *testing.T) {
	ctx := context.Background()
	cRepo := newMockContractorRepo()
	uRepo := newMockUserRepo()

	adminUserID := uuid.New()
	contractorID := uuid.New()
	cRepo.contractors[contractorID] = &entity.Contractor{
		ID:          contractorID,
		Name:        "Constructora Regia",
		AdminUserID: &adminUserID,
		IsActive:    true,
	}

	getMyUC := contractor.NewGetMyContractorCompanyUseCase(cRepo)
	addWorkerUC := contractor.NewAddContractorWorkerUseCase(cRepo, uRepo)
	listWorkersUC := contractor.NewListContractorWorkersUseCase(cRepo)
	toggleWorkerUC := contractor.NewToggleContractorWorkerStatusUseCase(cRepo)

	t.Run("getMyContractorCompany returns correct entity", func(t *testing.T) {
		c, err := getMyUC.Execute(ctx, adminUserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != contractorID {
			t.Errorf("expected contractor ID %s, got %s", contractorID, c.ID)
		}
	})

	t.Run("addWorker successfully registers worker", func(t *testing.T) {
		workerDetail, err := addWorkerUC.Execute(ctx, contractor.AddContractorWorkerInput{
			ContractorID: contractorID,
			FullName:     "Juan Perez Martinez",
			Email:        "juan.perez@regia.mx",
			Phone:        "8111223344",
			JobTitle:     "Soldador Calificado",
			CURP:         "PEMJ900101HNLNN01",
			IMSSNumber:   "12345678901",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if workerDetail.FullName != "Juan Perez Martinez" {
			t.Errorf("got %s", workerDetail.FullName)
		}

		workers, err := listWorkersUC.Execute(ctx, contractorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(workers) != 1 {
			t.Fatalf("expected 1 worker, got %d", len(workers))
		}

		// Toggle status
		err = toggleWorkerUC.Execute(ctx, contractorID, workerDetail.WorkerID, false)
		if err != nil {
			t.Fatalf("unexpected error toggling status: %v", err)
		}
	})
}
