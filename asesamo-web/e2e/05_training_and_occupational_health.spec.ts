import { test, expect } from '@playwright/test';
import { mockAuthenticatedSession } from './helpers';

test.describe('E2E: Capacitación STPS (DC-3) y Salud Ocupacional (NOM-030)', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedSession(page, 'ehs_manager');
  });

  test('Accede al módulo de Capacitación y constancias DC-3 en /capacitacion', async ({ page }) => {
    await page.goto('/capacitacion');
    await expect(page.getByRole('heading', { name: 'Capacitación, Adiestramiento y Habilidades Laborales' })).toBeVisible();
    await expect(page.locator('text=Constancias DC-3 Emitidas').first()).toBeVisible();
    await expect(page.locator('text=Catálogo de Cursos STPS').first()).toBeVisible();
  });

  test('Accede al módulo de Salud Ocupacional y Medicina del Trabajo en /salud-ocupacional', async ({ page }) => {
    await page.goto('/salud-ocupacional');
    await expect(page.getByRole('heading', { name: 'Salud Ocupacional & Dictámenes de Aptitud' })).toBeVisible();
    await expect(page.locator('text=Total Evaluaciones Clínicas').first()).toBeVisible();
    await expect(page.locator('text=Bitácora de Exámenes y Dictámenes Médicos Industriales').first()).toBeVisible();
  });
});
