import { test, expect } from '@playwright/test';
import { mockAuthenticatedSession } from './helpers';

test.describe('E2E: Inspecciones de Seguridad en Campo & Planes de Acción CAPA', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedSession(page, 'ehs_manager');
  });

  test('Muestra el panel de Inspecciones con métricas y botón de nueva auditoría', async ({ page }) => {
    await page.goto('/inspecciones');
    await expect(page.locator('text=Inspecciones & Auditorías EHS').first()).toBeVisible();
    await expect(page.locator('text=Nueva Inspección').first()).toBeVisible();
  });

  test('Permite ingresar al formulario de inspección en campo /inspecciones/nueva', async ({ page }) => {
    await page.goto('/inspecciones');
    await page.click('text=Nueva Inspección');
    await expect(page).toHaveURL('/inspecciones/nueva');
    await expect(page.getByRole('heading', { name: 'Nueva Inspección de Seguridad' })).toBeVisible();
  });
});
