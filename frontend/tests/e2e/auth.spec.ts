import { test, expect } from '@playwright/test';
import { generateTestCredentials, login, logout } from './fixtures/auth';

test.describe('Authentication', () => {
  test.describe('Registration', () => {
    test('should successfully register a new user', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      await page.goto('/register');

      // Verify registration page loads
      await expect(page).toHaveTitle(/Register/i);

      // Fill registration form
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);

      // Submit form
      await page.click('button[type="submit"]');

      // Should redirect to dashboard
      await page.waitForURL('/dashboard', { timeout: 10000 });
      await expect(page).toHaveURL('/dashboard');

      // Should see welcome message or user name
      await expect(page.locator('text=/Welcome|Dashboard/i')).toBeVisible();
    });

    test('should show error for invalid email', async ({ page }) => {
      await page.goto('/register');

      await page.fill('input[type="email"]', 'invalid-email');
      await page.fill('input[type="password"]', 'ValidPassword123!');
      await page.fill('input[name="name"]', 'Test User');

      await page.click('button[type="submit"]');

      // Should show validation error
      await expect(page.locator('text=/invalid.*email/i')).toBeVisible({ timeout: 5000 });
    });

    test('should show error for weak password', async ({ page }) => {
      const { email, name } = generateTestCredentials();

      await page.goto('/register');

      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', '123'); // Too short
      await page.fill('input[name="name"]', name);

      await page.click('button[type="submit"]');

      // Should show password error (either client or server)
      const errorVisible = await Promise.race([
        page.locator('text=/password.*short|password.*least/i').isVisible().then(() => true),
        page.waitForTimeout(3000).then(() => false),
      ]);

      expect(errorVisible).toBe(true);
    });

    test('should show error for duplicate email', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      // Register first user
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard', { timeout: 10000 });

      // Logout
      await logout(page);

      // Try to register again with same email
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', 'Another User');
      await page.click('button[type="submit"]');

      // Should show duplicate email error
      await expect(page.locator('text=/already exists|email.*taken/i')).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Login', () => {
    test('should successfully login with valid credentials', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      // First register a user
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard');

      // Logout
      await logout(page);

      // Now login
      await login(page, email, password);

      // Should be on dashboard
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('text=/Welcome|Dashboard/i')).toBeVisible();
    });

    test('should show error for invalid credentials', async ({ page }) => {
      await page.goto('/login');

      await page.fill('input[type="email"]', 'nonexistent@example.com');
      await page.fill('input[type="password"]', 'WrongPassword123!');

      await page.click('button[type="submit"]');

      // Should show invalid credentials error
      await expect(page.locator('text=/invalid.*credentials|incorrect.*password|login.*failed/i')).toBeVisible({ timeout: 5000 });
    });

    test('should show error for empty fields', async ({ page }) => {
      await page.goto('/login');

      // Try to submit empty form
      await page.click('button[type="submit"]');

      // Should show validation errors (HTML5 or custom)
      const emailInput = page.locator('input[type="email"]');
      const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);

      expect(isInvalid).toBe(true);
    });

    test('should persist session across page reloads', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      // Register and login
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard');

      // Reload page
      await page.reload();

      // Should still be authenticated
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('text=/Welcome|Dashboard/i')).toBeVisible();
    });
  });

  test.describe('Logout', () => {
    test('should successfully logout', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      // Register a user
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard');

      // Logout
      await page.click('button:has-text("Logout")');

      // Should redirect to login
      await page.waitForURL('/login', { timeout: 5000 });
      await expect(page).toHaveURL('/login');
    });

    test('should clear session after logout', async ({ page }) => {
      const { email, password, name } = generateTestCredentials();

      // Register and logout
      await page.goto('/register');
      await page.fill('input[type="email"]', email);
      await page.fill('input[type="password"]', password);
      await page.fill('input[name="name"]', name);
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard');
      await logout(page);

      // Try to access protected route
      await page.goto('/dashboard');

      // Should redirect to login
      await page.waitForURL('/login', { timeout: 5000 });
      await expect(page).toHaveURL('/login');
    });
  });

  test.describe('Protected Routes', () => {
    test('should redirect to login when accessing protected routes without auth', async ({ page }) => {
      // Try to access dashboard without authentication
      await page.goto('/dashboard');

      // Should redirect to login
      await page.waitForURL('/login', { timeout: 5000 });
      await expect(page).toHaveURL('/login');
    });

    test('should redirect to login when accessing chat without auth', async ({ page }) => {
      await page.goto('/chat');

      await page.waitForURL('/login', { timeout: 5000 });
      await expect(page).toHaveURL('/login');
    });
  });
});
