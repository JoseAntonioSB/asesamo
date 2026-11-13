package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type EPPRepository struct {
	db *sql.DB
}

func NewEPPRepository(db *sql.DB) *EPPRepository {
	return &EPPRepository{db: db}
}

func (r *EPPRepository) CreateItem(ctx context.Context, item *entity.EPPItem) error {
	query := `
		INSERT INTO epp_catalog (
			id, tenant_id, code, name, category, normative_standard, lifespan_days, stock_quantity, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5::epp_category, $6, $7, $8, $9, $10, $11
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.TenantID, item.Code, item.Name, string(item.Category),
		item.NormativeStandard, item.LifespanDays, item.StockQuantity, item.IsActive, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al registrar artículo EPP: %w", err)
	}
	return nil
}

func (r *EPPRepository) ListItemCatalog(ctx context.Context, tenantID uuid.UUID) ([]entity.EPPItem, error) {
	query := `
		SELECT id, tenant_id, code, name, category, normative_standard, lifespan_days, stock_quantity, is_active, created_at, updated_at
		FROM epp_catalog
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar catálogo EPP: %w", err)
	}
	defer rows.Close()

	var list []entity.EPPItem
	for rows.Next() {
		var itm entity.EPPItem
		var std sql.NullString
		if err := rows.Scan(
			&itm.ID, &itm.TenantID, &itm.Code, &itm.Name, &itm.Category, &std,
			&itm.LifespanDays, &itm.StockQuantity, &itm.IsActive, &itm.CreatedAt, &itm.UpdatedAt,
		); err != nil {
			return nil, err
		}
		itm.NormativeStandard = std.String
		list = append(list, itm)
	}
	return list, nil
}

func (r *EPPRepository) GetItemByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EPPItem, error) {
	query := `
		SELECT id, tenant_id, code, name, category, normative_standard, lifespan_days, stock_quantity, is_active, created_at, updated_at
		FROM epp_catalog
		WHERE tenant_id = $1 AND id = $2
	`
	itm := &entity.EPPItem{}
	var std sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&itm.ID, &itm.TenantID, &itm.Code, &itm.Name, &itm.Category, &std,
		&itm.LifespanDays, &itm.StockQuantity, &itm.IsActive, &itm.CreatedAt, &itm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener EPP: %w", err)
	}
	itm.NormativeStandard = std.String
	return itm, nil
}

func (r *EPPRepository) RecordDelivery(ctx context.Context, delivery *entity.EPPDelivery) error {
	query := `
		INSERT INTO epp_deliveries (
			id, tenant_id, epp_item_id, worker_id, contractor_id, delivery_date,
			expiration_date, status, delivered_by_user_id, signature_data, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8::epp_delivery_status, $9, $10, $11,
			$12, $13
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		delivery.ID, delivery.TenantID, delivery.EPPItemID, delivery.WorkerID, delivery.ContractorID, delivery.DeliveryDate,
		delivery.ExpirationDate, string(delivery.Status), delivery.DeliveredByUserID, delivery.SignatureData, delivery.Notes,
		delivery.CreatedAt, delivery.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al registrar entrega de EPP: %w", err)
	}
	return nil
}

func (r *EPPRepository) ListDeliveries(ctx context.Context, tenantID uuid.UUID, filter repository.EPPDeliveryFilter) ([]entity.EPPDelivery, int, error) {
	baseQuery := `
		SELECT id, tenant_id, epp_item_id, worker_id, contractor_id, delivery_date,
		       expiration_date, status, delivered_by_user_id, signature_data, notes,
		       created_at, updated_at
		FROM epp_deliveries
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.WorkerID != nil {
		baseQuery += fmt.Sprintf(" AND worker_id = $%d", argIdx)
		args = append(args, *filter.WorkerID)
		argIdx++
	}
	if filter.ContractorID != nil {
		baseQuery += fmt.Sprintf(" AND contractor_id = $%d", argIdx)
		args = append(args, *filter.ContractorID)
		argIdx++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d::epp_delivery_status", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}

	baseQuery += " ORDER BY delivery_date DESC"
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar entregas de EPP: %w", err)
	}
	defer rows.Close()

	var list []entity.EPPDelivery
	for rows.Next() {
		var d entity.EPPDelivery
		var sig, notes sql.NullString
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.EPPItemID, &d.WorkerID, &d.ContractorID, &d.DeliveryDate,
			&d.ExpirationDate, &d.Status, &d.DeliveredByUserID, &sig, &notes,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		d.SignatureData = sig.String
		d.Notes = notes.String
		list = append(list, d)
	}

	return list, len(list), nil
}

func (r *EPPRepository) UpdateDeliveryStatus(ctx context.Context, tenantID, deliveryID uuid.UUID, status entity.EPPDeliveryStatus, notes string) error {
	query := `
		UPDATE epp_deliveries
		SET status = $1::epp_delivery_status,
		    notes = CASE WHEN $2 != '' THEN $2 ELSE notes END,
		    updated_at = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	res, err := r.db.ExecContext(ctx, query, string(status), notes, tenantID, deliveryID)
	if err != nil {
		return fmt.Errorf("error al actualizar estatus de EPP: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("entrega de EPP no encontrada")
	}
	return nil
}
