package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/contractor"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contractor Repo Mock for Handler
type handlerMockContractorRepo struct {
	contractors      map[uuid.UUID]*entity.Contractor
	contractorsByRFC map[string]*entity.Contractor
	plantContractors map[string]*entity.PlantContractor
	workers          map[uuid.UUID]*entity.ContractorWorker
	users            map[uuid.UUID]*entity.User
}

func newHandlerMockContractorRepo() *handlerMockContractorRepo {
	return &handlerMockContractorRepo{
		contractors:      make(map[uuid.UUID]*entity.Contractor),
		contractorsByRFC: make(map[string]*entity.Contractor),
		plantContractors: make(map[string]*entity.PlantContractor),
		workers:          make(map[uuid.UUID]*entity.ContractorWorker),
		users:            make(map[uuid.UUID]*entity.User),
	}
}

func (m *handlerMockContractorRepo) CreateContractor(ctx context.Context, c *entity.Contractor) error {
	m.contractors[c.ID] = c
	if c.RFC != nil {
		m.contractorsByRFC[*c.RFC] = c
	}
	return nil
}

func (m *handlerMockContractorRepo) GetContractorByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error) {
	c, ok := m.contractors[id]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (m *handlerMockContractorRepo) GetContractorByRFC(ctx context.Context, rfc string) (*entity.Contractor, error) {
	c, ok := m.contractorsByRFC[rfc]
	if !ok {
		return nil, domainErrors.ErrContractorNotFound
	}
	return c, nil
}

func (m *handlerMockContractorRepo) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	for _, c := range m.contractors {
		if c.AdminUserID != nil && *c.AdminUserID == adminUserID {
			return c, nil
		}
	}
	return nil, domainErrors.ErrContractorNotFound
}

func (m *handlerMockContractorRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]repository.ContractorWithDetails, error) {
	var list []repository.ContractorWithDetails
	for _, pc := range m.plantContractors {
		if pc.PlantID == plantID {
			c := m.contractors[pc.ContractorID]
			if c != nil {
				list = append(list, repository.ContractorWithDetails{
					ID:             c.ID,
					Name:           c.Name,
					RFC:            c.RFC,
					PlantStatus:    pc.Status,
					IsActive:       c.IsActive,
				})
			}
		}
	}
	return list, nil
}

func (m *handlerMockContractorRepo) CreatePlantContractor(ctx context.Context, pc *entity.PlantContractor) error {
	m.plantContractors[pc.PlantID.String()+":"+pc.ContractorID.String()] = pc
	return nil
}

func (m *handlerMockContractorRepo) UpdatePlantContractorStatus(ctx context.Context, plantID, contractorID uuid.UUID, status entity.PlantContractorStatus) error {
	key := plantID.String() + ":" + contractorID.String()
	pc, ok := m.plantContractors[key]
	if !ok {
		return domainErrors.ErrContractorNotFound
	}
	pc.Status = status
	return nil
}

func (m *handlerMockContractorRepo) ListWorkersByContractor(ctx context.Context, contractorID uuid.UUID) ([]repository.ContractorWorkerDetail, error) {
	var list []repository.ContractorWorkerDetail
	for _, w := range m.workers {
		if w.ContractorID == contractorID {
			u := m.users[w.UserID]
			fn := "Trabajador"
			em := ""
			if u != nil {
				fn = u.FullName
				em = u.Email
			}
			list = append(list, repository.ContractorWorkerDetail{
				WorkerID:     w.UserID,
				ContractorID: w.ContractorID,
				FullName:     fn,
				Email:        em,
				IsActive:     w.IsActive,
			})
		}
	}
	return list, nil
}

func (m *handlerMockContractorRepo) AddWorker(ctx context.Context, workerUser *entity.User, cw *entity.ContractorWorker) error {
	m.users[workerUser.ID] = workerUser
	m.workers[cw.ID] = cw
	return nil
}

func (m *handlerMockContractorRepo) ToggleWorkerStatus(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error {
	for _, w := range m.workers {
		if w.ContractorID == contractorID && w.UserID == workerUserID {
			w.IsActive = isActive
			return nil
		}
	}
	return domainErrors.ErrWorkerNotFound
}

func (m *handlerMockContractorRepo) CreateUserWithPassword(ctx context.Context, user *entity.User, passwordHash string) error {
	m.users[user.ID] = user
	return nil
}

// User repo mock for handler
type handlerMockUserRepo struct {
	users map[string]*entity.User
}

func (m *handlerMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domainErrors.ErrUserNotFound
}

func (m *handlerMockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, domainErrors.ErrUserNotFound
	}
	return u, nil
}

func (m *handlerMockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

func (m *handlerMockUserRepo) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	return nil
}

func (m *handlerMockUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, fullName string, phone *string) error {
	return nil
}

func (m *handlerMockUserRepo) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarKey string) error {
	return nil
}

func (m *handlerMockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

// Plant repo mock for handler
type handlerMockPlantRepo struct {
	plants map[uuid.UUID]*entity.Plant
}

func (m *handlerMockPlantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	p, ok := m.plants[id]
	if !ok {
		return nil, domainErrors.ErrPlantNotFound
	}
	return p, nil
}

func (m *handlerMockPlantRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error) {
	return nil, nil
}

func TestContractorHandler(t *testing.T) {
	cRepo := newHandlerMockContractorRepo()
	uRepo := &handlerMockUserRepo{users: make(map[string]*entity.User)}
	pRepo := &handlerMockPlantRepo{plants: make(map[uuid.UUID]*entity.Plant)}

	plantID := uuid.New()
	tenantID := uuid.New()
	adminUserID := uuid.New()
	pRepo.plants[plantID] = &entity.Plant{
		ID:       plantID,
		TenantID: tenantID,
		Name:     "Planta Guadalajara",
	}

	inviteUC := contractor.NewInviteContractorUseCase(cRepo, uRepo, pRepo)
	listUC := contractor.NewListPlantContractorsUseCase(cRepo, pRepo)
	updateStatusUC := contractor.NewUpdateContractorStatusUseCase(cRepo)
	getMyUC := contractor.NewGetMyContractorCompanyUseCase(cRepo)
	listWorkersUC := contractor.NewListContractorWorkersUseCase(cRepo)
	addWorkerUC := contractor.NewAddContractorWorkerUseCase(cRepo, uRepo)
	toggleStatusUC := contractor.NewToggleContractorWorkerStatusUseCase(cRepo)

	h := handler.NewContractorHandler(
		inviteUC,
		listUC,
		updateStatusUC,
		getMyUC,
		listWorkersUC,
		addWorkerUC,
		toggleStatusUC,
	)

	r := chi.NewRouter()
	r.Post("/api/v1/contractors/invite", h.InviteContractor)
	r.Get("/api/v1/contractors", h.ListPlantContractors)
	r.Patch("/api/v1/contractors/{id}/status", h.UpdateContractorStatus)
	r.Get("/api/v1/contractors/my-company", h.GetMyCompany)
	r.Get("/api/v1/contractors/workers", h.ListWorkers)
	r.Post("/api/v1/contractors/workers", h.AddWorker)
	r.Patch("/api/v1/contractors/workers/{id}/status", h.ToggleWorkerStatus)

	t.Run("POST /api/v1/contractors/invite creates company", func(t *testing.T) {
		body := map[string]interface{}{
			"plant_id":        plantID.String(),
			"name":            "Mantenimiento Industrial SA de CV",
			"rfc":             "MIS850101XYZ",
			"repse_number":    "REPSE-888-999",
			"contractor_type": "company",
			"admin_email":     "admin@mantenimiento.com",
			"admin_full_name": "Ing. Roberto Garza",
		}
		raw, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/contractors/invite", bytes.NewReader(raw))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		req = req.WithContext(tenant.WithUserID(req.Context(), adminUserID))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GET /api/v1/contractors lists plant contractors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/contractors?plant_id="+plantID.String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var res struct {
			Data []map[string]interface{} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Len(t, res.Data, 1)
	})
}
