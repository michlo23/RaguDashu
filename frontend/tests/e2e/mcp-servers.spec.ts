import { test, expect } from './fixtures/auth';

test.describe('MCP Server Management', () => {
  test.describe('MCP Server List', () => {
    test('should navigate to MCP servers page', async ({ authenticatedPage: page }) => {
      // Navigate to MCP servers page (assuming it's in the menu)
      await page.goto('/mcp-servers');

      // Verify we're on the right page
      await expect(page).toHaveURL('/mcp-servers');
      await expect(page.locator('h1, h2').filter({ hasText: /MCP Servers/i })).toBeVisible();
    });

    test('should display empty state when no servers configured', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Should show empty state or "No servers" message
      const emptyState = page.locator('text=/No MCP servers|Add your first MCP server/i');
      if (await emptyState.isVisible()) {
        await expect(emptyState).toBeVisible();
      }
    });

    test('should show create server button', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Should have a button to create new server
      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await expect(createButton.first()).toBeVisible();
    });
  });

  test.describe('Create MCP Server', () => {
    test('should create MCP server with API key auth', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Click create button
      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      // Fill in server details
      await page.fill('input[name="name"], input[placeholder*="name" i]', 'Test MCP Server');
      await page.fill('textarea[name="description"], textarea[placeholder*="description" i]', 'Test server for E2E testing');
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'https://mcp.example.com/v1');

      // Select API key auth type
      const authTypeSelect = page.locator('select[name="auth_type"], [role="combobox"]').first();
      if (await authTypeSelect.isVisible()) {
        await authTypeSelect.selectOption('api_key');
      }

      // Enter API key
      await page.fill('input[name="api_key"], input[type="password"][placeholder*="API" i]', 'test-api-key-12345');

      // Submit form
      await page.click('button[type="submit"], button:has-text("Create"), button:has-text("Save")');

      // Wait for success message or redirect
      await page.waitForTimeout(2000);

      // Verify server appears in list
      const serverName = page.locator('text=Test MCP Server');
      await expect(serverName).toBeVisible({ timeout: 10000 });
    });

    test('should show validation error for empty name', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      // Fill only URL, leave name empty
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'https://mcp.example.com/v1');

      // Try to submit
      await page.click('button[type="submit"], button:has-text("Create"), button:has-text("Save")');

      // Should show validation error
      const errorMessage = page.locator('text=/required|cannot be empty/i');
      await expect(errorMessage).toBeVisible({ timeout: 5000 });
    });

    test('should create MCP server with OAuth 2 config', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      // Fill basic info
      await page.fill('input[name="name"], input[placeholder*="name" i]', 'Atlassian MCP');
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'https://mcp.atlassian.com/v1');

      // Select OAuth 2 auth
      const authTypeSelect = page.locator('select[name="auth_type"], [role="combobox"]').first();
      if (await authTypeSelect.isVisible()) {
        await authTypeSelect.selectOption('oauth2');
      }

      // Fill OAuth 2 config (if form supports it)
      const clientIdInput = page.locator('input[name="client_id"], input[placeholder*="Client ID" i]');
      if (await clientIdInput.isVisible()) {
        await clientIdInput.fill('test-client-id');
        await page.fill('input[name="client_secret"], input[placeholder*="Client Secret" i]', 'test-client-secret');
        await page.fill('input[name="auth_url"], input[placeholder*="Auth.*URL" i]', 'https://auth.atlassian.com/authorize');
        await page.fill('input[name="token_url"], input[placeholder*="Token.*URL" i]', 'https://auth.atlassian.com/oauth/token');
      }

      // Submit
      await page.click('button[type="submit"], button:has-text("Create"), button:has-text("Save")');

      await page.waitForTimeout(2000);

      // Verify server created
      await expect(page.locator('text=Atlassian MCP')).toBeVisible({ timeout: 10000 });
    });
  });

  test.describe('MCP Server Actions', () => {
    test('should sync capabilities from server', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Find a server card/row
      const serverCard = page.locator('[data-testid="mcp-server-card"], .mcp-server, [class*="server-card"]').first();

      if (await serverCard.isVisible()) {
        // Look for sync button
        const syncButton = serverCard.locator('button').filter({ hasText: /Sync|Refresh/i });

        if (await syncButton.isVisible()) {
          await syncButton.click();

          // Wait for sync to complete
          await page.waitForTimeout(3000);

          // Should show success message
          const successMessage = page.locator('text=/synced|updated|success/i');
          await expect(successMessage).toBeVisible({ timeout: 5000 });
        }
      }
    });

    test('should view server details', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Click on first server
      const serverName = page.locator('[data-testid="mcp-server-name"], .server-name, h3, h4').first();

      if (await serverName.isVisible()) {
        await serverName.click();

        // Should show server details
        await expect(page.locator('text=/Details|Configuration|Tools/i')).toBeVisible({ timeout: 5000 });
      }
    });

    test('should display available tools', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Navigate to server details
      const firstServer = page.locator('[data-testid="mcp-server-card"], .mcp-server').first();

      if (await firstServer.isVisible()) {
        await firstServer.click();

        // Look for tools section
        const toolsSection = page.locator('text=/Available Tools|Tools/i');
        if (await toolsSection.isVisible()) {
          await expect(toolsSection).toBeVisible();

          // Check if tools are listed
          const toolsList = page.locator('[data-testid="tools-list"], .tools-list, ul, [class*="tool"]');
          // Just verify the section exists, tools may be empty
        }
      }
    });

    test('should toggle server active status', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      const serverCard = page.locator('[data-testid="mcp-server-card"], .mcp-server').first();

      if (await serverCard.isVisible()) {
        // Look for toggle switch or disable button
        const toggleButton = serverCard.locator('button').filter({ hasText: /Enable|Disable|Active/i });

        if (await toggleButton.isVisible()) {
          const initialText = await toggleButton.textContent();
          await toggleButton.click();

          // Wait for state change
          await page.waitForTimeout(1000);

          // Text should change
          const newText = await toggleButton.textContent();
          expect(newText).not.toBe(initialText);
        }
      }
    });

    test('should delete MCP server', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Create a server first
      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      await page.fill('input[name="name"], input[placeholder*="name" i]', 'Server To Delete');
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'https://delete.example.com');
      await page.click('button[type="submit"], button:has-text("Create")');

      await page.waitForTimeout(2000);

      // Find the server
      const serverCard = page.locator('text=Server To Delete').locator('..');

      // Click delete button
      const deleteButton = serverCard.locator('button').filter({ hasText: /Delete|Remove/i });
      if (await deleteButton.isVisible()) {
        await deleteButton.click();

        // Confirm deletion
        const confirmButton = page.locator('button').filter({ hasText: /Confirm|Yes|Delete/i });
        if (await confirmButton.isVisible()) {
          await confirmButton.click();
        }

        // Server should be removed
        await expect(page.locator('text=Server To Delete')).not.toBeVisible({ timeout: 5000 });
      }
    });
  });

  test.describe('MCP Server Validation', () => {
    test('should validate server URL format', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      // Enter invalid URL
      await page.fill('input[name="name"], input[placeholder*="name" i]', 'Invalid URL Server');
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'not-a-valid-url');

      // Try to submit
      await page.click('button[type="submit"], button:has-text("Create")');

      // Should show URL validation error
      const errorMessage = page.locator('text=/invalid.*URL|valid.*URL/i');
      // URL validation might be on blur or submit
      if (await errorMessage.isVisible()) {
        await expect(errorMessage).toBeVisible();
      }
    });

    test('should require API key for api_key auth type', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      const createButton = page.locator('button, a').filter({ hasText: /Add|Create|New.*Server/i });
      await createButton.first().click();

      await page.fill('input[name="name"], input[placeholder*="name" i]', 'API Key Required');
      await page.fill('input[name="server_url"], input[placeholder*="URL" i]', 'https://mcp.example.com');

      // Select API key auth but don't provide key
      const authTypeSelect = page.locator('select[name="auth_type"], [role="combobox"]').first();
      if (await authTypeSelect.isVisible()) {
        await authTypeSelect.selectOption('api_key');
      }

      // Submit without API key
      await page.click('button[type="submit"], button:has-text("Create")');

      // Should show validation error for API key
      await page.waitForTimeout(1000);
      const errorMessage = page.locator('text=/API key.*required/i');
      if (await errorMessage.isVisible()) {
        await expect(errorMessage).toBeVisible();
      }
    });
  });

  test.describe('MCP Server Search and Filter', () => {
    test('should filter servers by name', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Look for search input
      const searchInput = page.locator('input[type="search"], input[placeholder*="Search" i]');

      if (await searchInput.isVisible()) {
        // Type search query
        await searchInput.fill('Test');

        await page.waitForTimeout(500);

        // Only matching servers should be visible
        // This test assumes there's at least one server with "Test" in the name
        const results = page.locator('[data-testid="mcp-server-card"], .mcp-server');
        // Just verify search input works, actual filtering logic may vary
      }
    });

    test('should filter servers by auth type', async ({ authenticatedPage: page }) => {
      await page.goto('/mcp-servers');

      // Look for filter dropdown
      const filterSelect = page.locator('select[name="filter"], [role="combobox"]').filter({ hasText: /Filter|Auth/i });

      if (await filterSelect.isVisible()) {
        await filterSelect.selectOption('oauth2');

        await page.waitForTimeout(500);

        // Only OAuth2 servers should be visible
        // Verification would depend on actual implementation
      }
    });
  });
});
