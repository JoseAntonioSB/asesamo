import { test, expect } from '@playwright/test';
import { mockAuthenticatedSession } from './helpers';

test.describe('E2E: Centro de Comando de Evacuación y Pase de Lista en Punto de Reunión', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedSession(page, 'ehs_manager');
  });

  test('Muestra el centro de comando de evacuación en /evacuacion', async ({ page }) => {
    await page.goto('/evacuacion');
    await expect(page.locator('text=Centro de Comando de Evacuación').first()).toBeVisible();
    await expect(page.locator('text=Condición Normal').first()).toBeVisible();
    await expect(page.locator('text=Activar Alarma de Evacuación').first()).toBeVisible();
  });
});
