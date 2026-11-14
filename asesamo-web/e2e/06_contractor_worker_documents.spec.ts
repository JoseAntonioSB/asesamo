import { test, expect } from '@playwright/test';
import { mockAuthenticatedSession } from './helpers';

test.describe('E2E: Asignación de Documentos a Trabajadores por Contratista', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedSession(page, 'contractor_admin');

    // Intercept document types catalog
    await page.route('**/api/v1/documents/types', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'dt-1',
              code: 'DC3',
              name: 'Constancia de Capacitación DC-3',
              applies_to: 'worker',
              is_required_default: true,
              default_validity_days: 365,
            },
            {
              id: 'dt-2',
              code: 'MEDICAL',
              name: 'Certificado Médico de Aptitud',
              applies_to: 'worker',
              is_required_default: true,
              default_validity_days: 365,
            },
            {
              id: 'dt-3',
              code: 'REPSE',
              name: 'Registro REPSE',
              applies_to: 'contractor',
              is_required_default: true,
              default_validity_days: 365,
            },
          ],
        }),
      });
    });

    // Intercept contractor company query
    await page.route('**/api/v1/contractors/my-company', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: 'c-1',
            name: 'Mantenimiento y Montajes del Norte S.A. de C.V.',
            rfc: 'MMN180315AB1',
            repse_number: 'AR28911/2021',
            plant_status: 'approved',
          },
        }),
      });
    });

    // Intercept contractor crew list
    await page.route('**/api/v1/contractors/*/workers', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              worker_id: 'w-1',
              user_id: 'w-1',
              full_name: 'Carlos Mendoza',
              email: 'carlos.mendoza@mantenimientonorte.com',
              job_title: 'Soldador Especialista',
              curp: 'MEDC880415HDFRNR02',
              imss_number: '23881234567',
              is_active: true,
              documents_count: 2,
            },
            {
              worker_id: 'w-2',
              user_id: 'w-2',
              full_name: 'Juan Perez',
              email: 'juan.perez@mantenimientonorte.com',
              job_title: 'Electricista Industrial',
              curp: 'PERJ900101HDFRNR01',
              imss_number: '12345678901',
              is_active: true,
              documents_count: 1,
            },
          ],
        }),
      });
    });

    // Intercept my-documents queries
    await page.route('**/api/v1/documents/my-documents*', async (route) => {
      const url = new URL(route.request().url());
      const workerId = url.searchParams.get('worker_id');

      if (workerId === 'w-1') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              {
                id: 'doc-101',
                file_name: 'DC3_Trabajos_Caliente_CarlosMendoza.pdf',
                document_type_code: 'DC3',
                document_type_name: 'Constancia de Capacitación DC-3',
                applies_to: 'worker',
                status: 'approved',
                issued_at: '2026-01-15',
                expires_at: '2027-01-15',
                days_until_expiry: 300,
                created_at: '2026-01-16T10:00:00Z',
              },
            ],
          }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: [
              {
                id: 'doc-contractor-1',
                file_name: 'REPSE_Vigente_2026.pdf',
                document_type_code: 'REPSE',
                document_type_name: 'Registro REPSE',
                applies_to: 'contractor',
                status: 'approved',
                issued_at: '2026-01-01',
                expires_at: '2027-01-01',
                days_until_expiry: 290,
                created_at: '2026-01-02T10:00:00Z',
              },
            ],
          }),
        });
      }
    });

    // Intercept upload endpoint
    await page.route('**/api/v1/documents/upload', async (route) => {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: 'doc-new-1',
            file_name: 'Constancia_DC3_Nueva.pdf',
            document_type_code: 'DC3',
            document_type_name: 'Constancia de Capacitación DC-3',
            status: 'pending_review',
            created_at: new Date().toISOString(),
          },
        }),
      });
    });
  });

  test('Padrón de contratistas muestra cuadrilla y enlace a expediente individual de cada trabajador', async ({ page }) => {
    await page.goto('/contratistas');
    await expect(page.getByRole('heading', { name: 'Padrón de Contratistas' })).toBeVisible();
    await expect(page.locator('text=Personal en Cuadrilla').first()).toBeVisible();

    // Verify workers are listed with their jobs and CURP
    await expect(page.locator('text=Carlos Mendoza').first()).toBeVisible();
    await expect(page.locator('text=Soldador Especialista').first()).toBeVisible();
    await expect(page.locator('text=Juan Perez').first()).toBeVisible();

    // Verify Expediente buttons
    const expedienteLinks = page.locator('a[href*="/documentos?worker_id="]');
    await expect(expedienteLinks.first()).toBeVisible();
  });

  test('Navegación al expediente de un trabajador filtra sus documentos individuales', async ({ page }) => {
    await page.goto('/documentos?worker_id=w-1');

    // Verify the Active Worker Banner appears with worker details
    await expect(page.locator('text=Expediente del Trabajador: Carlos Mendoza')).toBeVisible();
    await expect(page.locator('text=Filtro Individual')).toBeVisible();
    await expect(page.locator('text=Soldador Especialista')).toBeVisible();

    // Verify worker documents table renders worker's DC-3
    await expect(page.locator('text=DC3_Trabajos_Caliente_CarlosMendoza.pdf').first()).toBeVisible();
  });

  test('Modal de carga incluye selector obligatorio de trabajador al subir documentos de tipo worker (DC-3)', async ({ page }) => {
    await page.goto('/documentos?worker_id=w-1');

    // Open upload modal
    await page.click('button:has-text("Subir Documento")');
    await expect(page.getByRole('heading', { name: 'Carga de Documento de Cumplimiento' })).toBeVisible();

    // Verify Worker Selector is visible and preselected
    const workerSelect = page.locator('#workerSelect');
    await expect(workerSelect).toBeVisible();
    await expect(workerSelect).toHaveValue('w-1');

    // Change worker to Juan Perez
    await workerSelect.selectOption('w-2');
    await expect(workerSelect).toHaveValue('w-2');

    // Switch to REPSE (contractor document) and verify worker selector disappears
    await page.selectOption('#docType', 'REPSE');
    await expect(workerSelect).not.toBeVisible();

    // Switch back to DC-3 and verify worker selector reappears
    await page.selectOption('#docType', 'DC3');
    await expect(workerSelect).toBeVisible();
  });
});
