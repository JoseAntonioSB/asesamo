import { test, expect } from '@playwright/test';
import { TEST_USERS, mockAuthenticatedSession } from './helpers';

test.describe('E2E: Flujo de Autenticación & Redirección por Rol', () => {
  test('Usuario sin autenticar accediendo a /caseta es redirigido a /login con parámetro redirect', async ({ page }) => {
    await page.goto('/caseta');
    await expect(page).toHaveURL(/\/login\?redirect=%2Fcaseta/);
    await expect(page.getByRole('heading', { name: 'ASESAMO' })).toBeVisible();
    await expect(page.locator('input[type="email"]')).toBeVisible();
  });

  test('Usuario sin autenticar accediendo a /permisos es redirigido a /login con redirect', async ({ page }) => {
    await page.goto('/permisos');
    await expect(page).toHaveURL(/\/login\?redirect=%2Fpermisos/);
  });

  test('Usuario autenticado con rol guard accede a /caseta correctamente', async ({ page }) => {
    await mockAuthenticatedSession(page, 'guard');
    await page.goto('/caseta');
    await expect(page).toHaveURL('/caseta');
    await expect(page.locator('text=ASESAMO').first()).toBeVisible();
  });

  test('Usuario autenticado con rol ehs_manager accede a /permisos correctamente', async ({ page }) => {
    await mockAuthenticatedSession(page, 'ehs_manager');
    await page.goto('/permisos');
    await expect(page).toHaveURL('/permisos');
    await expect(page.getByRole('heading', { name: 'Permisos de Trabajo de Alto Riesgo' })).toBeVisible();
  });
});
