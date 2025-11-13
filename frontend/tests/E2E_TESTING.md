# E2E Testing with Playwright

This document provides comprehensive guidance on running and maintaining end-to-end (E2E) tests for the RAG Dashboard frontend.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Running Tests](#running-tests)
5. [Test Structure](#test-structure)
6. [Writing Tests](#writing-tests)
7. [Best Practices](#best-practices)
8. [CI/CD Integration](#cicd-integration)
9. [Troubleshooting](#troubleshooting)

---

## Overview

Our E2E tests use [Playwright](https://playwright.dev/), a modern cross-browser testing framework. The tests cover:

- **Authentication**: Registration, login, logout, session persistence
- **Document Management**: Upload, list, view, delete documents
- **Chat Functionality**: Send messages, receive responses, conversation management
- **Credentials**: Add, manage, test API keys
- **Error Handling**: Network errors, validation errors, API failures

### Test Coverage

- ✅ 50+ test scenarios
- ✅ Cross-browser testing (Chrome, Firefox, Safari, Mobile)
- ✅ Visual regression testing capabilities
- ✅ Network mocking for error scenarios
- ✅ Authenticated and unauthenticated flows

---

## Prerequisites

### System Requirements

- **Node.js**: 16.x or higher
- **npm**: 8.x or higher
- **Browsers**: Playwright will auto-install Chromium, Firefox, and WebKit

### Backend Requirements

Tests require a running backend API. Configure via:

```bash
# .env.test (create this file)
VITE_API_URL=http://localhost:8080
PLAYWRIGHT_TEST_BASE_URL=http://localhost:5173
```

---

## Installation

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Install Playwright Browsers

```bash
npx playwright install
```

This downloads Chromium, Firefox, and WebKit browsers.

### 3. Verify Installation

```bash
npx playwright --version
```

Expected output: `Version 1.40.1`

---

## Running Tests

### All Tests

```bash
npm run test:e2e
```

### Specific Test File

```bash
npm run test:e2e tests/e2e/auth.spec.ts
```

### Interactive UI Mode

Best for development and debugging:

```bash
npm run test:e2e:ui
```

Features:
- Watch tests in real-time
- Time-travel debugging
- Step through tests
- Inspect DOM

### Headed Mode (See Browser)

```bash
npm run test:e2e:headed
```

### Debug Mode

```bash
npm run test:e2e:debug
```

Pauses execution at each step. Use Playwright Inspector to:
- Step through test
- Pick locators
- Edit locators
- See actionability logs

### Specific Browser

```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
```

### Mobile Testing

```bash
npx playwright test --project="Mobile Chrome"
npx playwright test --project="Mobile Safari"
```

### Specific Test

```bash
npx playwright test -g "should successfully login"
```

### View Report

After test run:

```bash
npm run test:e2e:report
```

---

## Test Structure

```
frontend/
├── tests/
│   └── e2e/
│       ├── fixtures/
│       │   └── auth.ts           # Authentication fixtures & helpers
│       ├── auth.spec.ts          # Authentication tests
│       ├── documents.spec.ts     # Document management tests
│       ├── chat.spec.ts          # Chat functionality tests
│       └── credentials.spec.ts   # Credentials management tests
├── playwright.config.ts          # Playwright configuration
└── package.json                  # Scripts & dependencies
```

### Test File Naming

- `*.spec.ts` - Test specifications
- `*.test.ts` - Alternative (also supported)

### Fixture Organization

- `fixtures/auth.ts` - Authentication helpers
- `fixtures/test-data.ts` - Test data generators (if needed)

---

## Writing Tests

### Basic Test Structure

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test('should do something', async ({ page }) => {
    // Navigate
    await page.goto('/path');

    // Interact
    await page.fill('input[name="email"]', 'test@example.com');
    await page.click('button[type="submit"]');

    // Assert
    await expect(page).toHaveURL('/dashboard');
  });
});
```

### Using Authentication Fixture

```typescript
import { test, expect } from './fixtures/auth';

test('should access protected route', async ({ authenticatedPage: page }) => {
  // Already logged in!
  await page.goto('/dashboard');
  await expect(page).toHaveURL('/dashboard');
});
```

### Locator Strategies

**Priority Order:**

1. **User-facing attributes** (best):
   ```typescript
   page.locator('button:has-text("Submit")');
   page.locator('text=Welcome');
   page.locator('[aria-label="Close"]');
   ```

2. **Semantic HTML**:
   ```typescript
   page.locator('button[type="submit"]');
   page.locator('input[type="email"]');
   ```

3. **Test IDs** (most reliable):
   ```typescript
   page.locator('[data-testid="login-button"]');
   ```

4. **CSS/XPath** (last resort):
   ```typescript
   page.locator('.submit-button');
   ```

### Common Patterns

#### Wait for Navigation

```typescript
await Promise.all([
  page.waitForNavigation(),
  page.click('a[href="/profile"]')
]);
```

#### Wait for API Response

```typescript
await Promise.all([
  page.waitForResponse(resp => resp.url().includes('/api/documents')),
  page.click('button:has-text("Upload")')
]);
```

#### Handle Dialogs

```typescript
page.on('dialog', dialog => dialog.accept());
await page.click('button:has-text("Delete")');
```

#### Mock API Responses

```typescript
await page.route('**/api/chat/send', route => {
  route.fulfill({
    status: 200,
    body: JSON.stringify({ message: 'Mocked response' })
  });
});
```

#### File Upload

```typescript
const fileInput = page.locator('input[type="file"]');
await fileInput.setInputFiles({
  name: 'test.txt',
  mimeType: 'text/plain',
  buffer: Buffer.from('Test content')
});
```

#### Take Screenshot

```typescript
await page.screenshot({ path: 'screenshot.png' });
```

---

## Best Practices

### 1. Test Independence

✅ **Good**: Each test sets up its own data
```typescript
test('should create document', async ({ authenticatedPage }) => {
  const uniqueName = `doc-${Date.now()}.txt`;
  // Test uses unique data
});
```

❌ **Bad**: Tests depend on each other
```typescript
test('create document', async ({ page }) => {
  // Creates "test.txt"
});

test('delete document', async ({ page }) => {
  // Assumes "test.txt" exists
});
```

### 2. Use Appropriate Timeouts

```typescript
// Default timeout (good for most cases)
await expect(page.locator('text=Success')).toBeVisible();

// Custom timeout for slow operations
await expect(page.locator('text=Processed')).toBeVisible({ timeout: 30000 });

// Short timeout for "should not exist"
await expect(page.locator('text=Error')).not.toBeVisible({ timeout: 2000 });
```

### 3. Avoid Hard-Coded Waits

❌ **Bad**:
```typescript
await page.waitForTimeout(5000); // Flaky!
```

✅ **Good**:
```typescript
await page.waitForSelector('text=Loaded');
await expect(page.locator('text=Ready')).toBeVisible();
```

### 4. Use Page Object Model (For Complex Pages)

```typescript
// pages/LoginPage.ts
export class LoginPage {
  constructor(private page: Page) {}

  async login(email: string, password: string) {
    await this.page.fill('input[type="email"]', email);
    await this.page.fill('input[type="password"]', password);
    await this.page.click('button[type="submit"]');
  }
}

// In test
const loginPage = new LoginPage(page);
await loginPage.login('test@example.com', 'password');
```

### 5. Test User Flows, Not Implementation

✅ **Good**: Tests user journey
```typescript
test('user can register and upload document', async ({ page }) => {
  await register(page, email, password, name);
  await uploadDocument(page, 'test.pdf');
  await expect(page.locator('text=test.pdf')).toBeVisible();
});
```

❌ **Bad**: Tests internal state
```typescript
test('redux state updates on upload', async ({ page }) => {
  // Too implementation-specific
});
```

### 6. Clean Test Data

```typescript
test.afterEach(async ({ page }) => {
  // Cleanup test data
  await page.evaluate(() => localStorage.clear());
});
```

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: |
          cd frontend
          npm ci
          npx playwright install --with-deps

      - name: Start backend
        run: |
          cd backend
          go run cmd/server/main.go &
          sleep 10

      - name: Run E2E tests
        run: |
          cd frontend
          npm run test:e2e

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: frontend/playwright-report/
```

### Docker Container

```dockerfile
# Dockerfile.test
FROM mcr.microsoft.com/playwright:v1.40.1-focal

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .

CMD ["npm", "run", "test:e2e"]
```

Run tests in Docker:
```bash
docker build -f Dockerfile.test -t rag-dashboard-tests .
docker run rag-dashboard-tests
```

---

## Troubleshooting

### Tests Timing Out

**Symptoms**: Tests hang or timeout

**Solutions**:
```bash
# Increase timeout
npx playwright test --timeout=60000

# Check backend is running
curl http://localhost:8080/api/health

# Run in headed mode to see what's happening
npm run test:e2e:headed
```

### Selector Not Found

**Symptoms**: `Error: Locator.click: Target closed` or `Element not found`

**Solutions**:
```typescript
// Add explicit wait
await page.waitForSelector('button:has-text("Submit")');

// Check if element exists
const exists = await page.locator('button').isVisible();

// Use more specific selector
page.locator('form button[type="submit"]');
```

### Flaky Tests

**Symptoms**: Tests pass/fail randomly

**Solutions**:
```typescript
// Use waitFor instead of fixed timeout
await expect(page.locator('text=Success')).toBeVisible({ timeout: 10000 });

// Wait for network idle
await page.goto('/dashboard', { waitUntil: 'networkidle' });

// Retry failed tests
// In playwright.config.ts
retries: process.env.CI ? 2 : 0,
```

### Authentication Issues

**Symptoms**: Tests fail at login/register

**Solutions**:
```bash
# Clear test database
psql $DATABASE_URL -c "TRUNCATE users CASCADE;"

# Check API is accessible
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","name":"Test"}'

# Enable debug logging
DEBUG=pw:api npm run test:e2e
```

### Browser Installation Issues

**Symptoms**: `Executable doesn't exist`

**Solutions**:
```bash
# Reinstall browsers
npx playwright install --force

# Install system dependencies (Linux)
npx playwright install-deps

# Check installation
npx playwright install --dry-run
```

---

## Advanced Topics

### Visual Regression Testing

```typescript
// Take screenshot
await page.screenshot({ path: 'baseline.png' });

// Compare with baseline
await expect(page).toHaveScreenshot('dashboard.png');
```

### Parallel Execution

```typescript
// playwright.config.ts
workers: process.env.CI ? 4 : 2,
fullyParallel: true,
```

### Test Annotations

```typescript
test('slow test', async ({ page }) => {
  test.slow(); // Triples timeout
  // Long-running test
});

test('skip on firefox', async ({ page, browserName }) => {
  test.skip(browserName === 'firefox', 'Feature not supported');
  // Test code
});

test.fixme('known issue', async ({ page }) => {
  // Marked as known failure
});
```

### Custom Fixtures

```typescript
import { test as base } from '@playwright/test';

export const test = base.extend({
  testDocuments: async ({ page }, use) => {
    const docs = await createTestDocuments();
    await use(docs);
    await deleteTestDocuments(docs);
  }
});
```

---

## Useful Commands

```bash
# List all tests
npx playwright test --list

# Run tests matching pattern
npx playwright test --grep "auth"

# Run tests with tag
npx playwright test --grep @smoke

# Show trace viewer
npx playwright show-trace trace.zip

# Generate test code (record actions)
npx playwright codegen http://localhost:5173

# Update screenshots
npx playwright test --update-snapshots
```

---

## Resources

- [Playwright Documentation](https://playwright.dev/)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [API Reference](https://playwright.dev/docs/api/class-playwright)
- [Debugging Guide](https://playwright.dev/docs/debug)
- [CI/CD Examples](https://playwright.dev/docs/ci)

---

## Support

For issues or questions:
1. Check [Troubleshooting](#troubleshooting) section
2. Review [Playwright Docs](https://playwright.dev/)
3. Open issue in project repository
4. Contact development team

---

**Happy Testing! 🎭**
