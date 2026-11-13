import { test, expect } from '@playwright/test';
import { mockAuthenticatedSession } from './helpers';

test.describe('E2E: Ciclo de Vida de Permisos de Trabajo PTAR (Alturas, Caliente, Confinados)', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedSession(page, 'ehs_manager');
  });

  test('Muestra el Dashboard de Permisos con Bento Grid de indicadores STPS', async ({ page }) => {
    await page.goto('/permisos');
    await expect(page.getByRole('heading', { name: 'Permisos de Trabajo de Alto Riesgo' })).toBeVisible();
    await expect(page.locator('text=Trabajos de Alto Riesgo en Planta').first()).toBeVisible();
    await expect(page.locator('text=Nuevo Permiso PTAR').first()).toBeVisible();
  });

  test('Navegación al formulario de nuevo permiso /permisos/nuevo', async ({ page }) => {
    await page.goto('/permisos');
    await page.click('text=Nuevo Permiso PTAR');
    await expect(page).toHaveURL('/permisos/nuevo');
    await expect(page.getByRole('heading', { name: '1. Selección de Riesgo Regulado STPS' })).toBeVisible();
    await expect(page.locator('text=Trabajos en Alturas').first()).toBeVisible();
  });
});
