import { test as base, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

export type AuthFixtures = {
  authenticatedPage: Page;
};

/**
 * Custom fixture that provides an authenticated page
 */
export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // Register a test user or use existing credentials
    const testEmail = `test-${Date.now()}@example.com`;
    const testPassword = 'TestPassword123!';
    const testName = 'Test User';

    // Navigate to register page
    await page.goto('/register');

    // Fill registration form
    await page.fill('input[type="email"]', testEmail);
    await page.fill('input[type="password"]', testPassword);
    await page.fill('input[name="name"]', testName);

    // Submit registration
    await page.click('button[type="submit"]');

    // Wait for redirect to dashboard
    await page.waitForURL('/dashboard', { timeout: 10000 });

    // Verify authentication
    await expect(page).toHaveURL('/dashboard');

    // Use the authenticated page
    await use(page);

    // Cleanup: logout after test
    try {
      await page.click('button:has-text("Logout")', { timeout: 2000 });
    } catch {
      // Ignore if logout button not found
    }
  },
});

export { expect };

/**
 * Helper function to login with existing credentials
 */
export async function login(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('/dashboard', { timeout: 10000 });
}

/**
 * Helper function to register a new user
 */
export async function register(page: Page, email: string, password: string, name: string) {
  await page.goto('/register');
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);
  await page.fill('input[name="name"]', name);
  await page.click('button[type="submit"]');
  await page.waitForURL('/dashboard', { timeout: 10000 });
}

/**
 * Helper function to logout
 */
export async function logout(page: Page) {
  await page.click('button:has-text("Logout")');
  await page.waitForURL('/login', { timeout: 5000 });
}

/**
 * Generate unique test credentials
 */
export function generateTestCredentials() {
  const timestamp = Date.now();
  return {
    email: `test-${timestamp}@example.com`,
    password: `TestPass${timestamp}!`,
    name: `Test User ${timestamp}`,
  };
}
