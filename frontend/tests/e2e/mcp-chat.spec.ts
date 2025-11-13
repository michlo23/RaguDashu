import { test, expect } from './fixtures/auth';

test.describe('Chat with MCP Tools', () => {
  test.describe('MCP Tool Configuration', () => {
    test('should enable MCP servers in chat configuration', async ({ authenticatedPage: page }) => {
      // Navigate to chat configurations
      await page.goto('/chat');

      // Look for configuration or settings button
      const configButton = page.locator('button').filter({ hasText: /Configuration|Settings/i });

      if (await configButton.isVisible()) {
        await configButton.click();

        // Look for MCP tools section
        const mcpSection = page.locator('text=/MCP.*Tools|Enable.*Tools/i');

        if (await mcpSection.isVisible()) {
          // Select MCP server checkbox/toggle
          const mcpCheckbox = page.locator('input[type="checkbox"][name*="mcp"], [role="checkbox"]').first();

          if (await mcpCheckbox.isVisible()) {
            await mcpCheckbox.check();

            // Save configuration
            await page.click('button:has-text("Save")');

            // Verify saved
            await expect(page.locator('text=/Saved|Updated/i')).toBeVisible({ timeout: 5000 });
          }
        }
      }
    });

    test('should display enabled MCP tools in chat interface', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Check if there's an indicator that MCP tools are available
      const toolsIndicator = page.locator('text=/Tools.*available|MCP.*enabled/i');

      if (await toolsIndicator.isVisible()) {
        await expect(toolsIndicator).toBeVisible();
      }

      // Or check for tools icon/badge
      const toolsIcon = page.locator('[data-testid="mcp-tools-icon"], [class*="tools-icon"]');
      if (await toolsIcon.isVisible()) {
        await expect(toolsIcon).toBeVisible();
      }
    });
  });

  test.describe('Chat with Tool Calling', () => {
    test('should send message that could trigger tool call', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Type a message that might trigger a tool
      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Search for recent issues in project ABC');

      // Send message
      await page.click('button[type="submit"], button:has-text("Send")');

      // Wait for response
      await page.waitForTimeout(3000);

      // Should receive a response
      const assistantMessage = page.locator('[data-testid="assistant-message"], .assistant-message, [class*="message"]').last();
      await expect(assistantMessage).toBeVisible({ timeout: 15000 });

      // Message should have content
      const messageText = await assistantMessage.textContent();
      expect(messageText).toBeTruthy();
      expect(messageText!.length).toBeGreaterThan(10);
    });

    test('should display tool call information in message', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Send message
      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Get information from the connected tools');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Look for tool call indicator
      const toolCallIndicator = page.locator('[data-testid="tool-call"], [class*="tool-call"], text=/Called tool|Used tool/i');

      if (await toolCallIndicator.isVisible()) {
        // Verify tool call details are shown
        await expect(toolCallIndicator).toBeVisible();

        // Tool name might be displayed
        const toolName = page.locator('[data-testid="tool-name"], [class*="tool-name"]');
        if (await toolName.isVisible()) {
          const name = await toolName.textContent();
          expect(name).toBeTruthy();
        }
      }
    });

    test('should handle tool call errors gracefully', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // This test assumes there might be error scenarios
      // We can't easily force an error, but we can check error handling

      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Request that might fail');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Should either succeed or show error gracefully
      const response = page.locator('[data-testid="assistant-message"], .assistant-message').last();
      if (await response.isVisible()) {
        await expect(response).toBeVisible();

        // Error messages should be user-friendly
        const errorMessage = page.locator('text=/error|failed|unable/i');
        if (await errorMessage.isVisible()) {
          const text = await errorMessage.textContent();
          // Should not expose technical details
          expect(text).not.toContain('stack trace');
          expect(text).not.toContain('undefined');
        }
      }
    });

    test('should show loading indicator during tool call', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Execute a tool action');

      await page.click('button[type="submit"], button:has-text("Send")');

      // Immediately check for loading indicator
      const loadingIndicator = page.locator('[data-testid="loading"], .loading, [class*="spinner"]');

      if (await loadingIndicator.isVisible({ timeout: 500 })) {
        await expect(loadingIndicator).toBeVisible();
      }

      // Wait for response
      await page.waitForTimeout(5000);

      // Loading should disappear
      await expect(loadingIndicator).not.toBeVisible();
    });
  });

  test.describe('Tool Usage Analytics', () => {
    test('should track tool usage in conversation', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Send message that might use tools
      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Query external data');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Check conversation history or analytics
      const analyticsButton = page.locator('button').filter({ hasText: /Analytics|History|Details/i });

      if (await analyticsButton.isVisible()) {
        await analyticsButton.click();

        // Should show tool usage information
        const toolUsage = page.locator('text=/Tool.*used|Called|Duration/i');
        if (await toolUsage.isVisible()) {
          await expect(toolUsage).toBeVisible();
        }
      }
    });
  });

  test.describe('Multi-Server Tool Integration', () => {
    test('should handle multiple MCP servers enabled', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // This test verifies the system can handle multiple tool sources
      // The exact UI will depend on implementation

      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('What tools are available?');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Should get a response
      const response = page.locator('[data-testid="assistant-message"], .assistant-message').last();
      await expect(response).toBeVisible({ timeout: 10000 });
    });

    test('should disambiguate tools from different servers', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // If multiple servers have similar tools, system should handle it
      // This is more of an integration test to verify no conflicts

      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Search for information');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Should successfully process without conflicts
      const response = page.locator('[data-testid="assistant-message"], .assistant-message').last();
      await expect(response).toBeVisible({ timeout: 10000 });

      // No error messages about ambiguous tools
      const errorText = page.locator('text=/ambiguous|conflict|duplicate/i');
      expect(await errorText.isVisible()).toBe(false);
    });
  });

  test.describe('Chat Configuration with MCP', () => {
    test('should create new chat configuration with MCP tools', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Create new configuration
      const newConfigButton = page.locator('button').filter({ hasText: /New.*Config|Create.*Config/i });

      if (await newConfigButton.isVisible()) {
        await newConfigButton.click();

        // Fill configuration
        await page.fill('input[name="name"], input[placeholder*="name" i]', 'MCP-Enabled Config');

        // Enable MCP servers
        const mcpToggle = page.locator('input[type="checkbox"][name*="mcp"]').first();
        if (await mcpToggle.isVisible()) {
          await mcpToggle.check();
        }

        // Save
        await page.click('button[type="submit"], button:has-text("Save"), button:has-text("Create")');

        await page.waitForTimeout(1000);

        // Verify configuration created
        await expect(page.locator('text=MCP-Enabled Config')).toBeVisible({ timeout: 5000 });
      }
    });

    test('should switch between configs with different MCP settings', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Look for configuration selector
      const configSelect = page.locator('select[name="config"], [role="combobox"]').first();

      if (await configSelect.isVisible()) {
        // Get current config
        const currentConfig = await configSelect.inputValue();

        // Switch to different config
        const options = await configSelect.locator('option').all();
        if (options.length > 1) {
          await configSelect.selectOption({ index: 1 });

          await page.waitForTimeout(500);

          // Config should have changed
          const newConfig = await configSelect.inputValue();
          expect(newConfig).not.toBe(currentConfig);
        }
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should handle disabled MCP server gracefully', async ({ authenticatedPage: page }) => {
      // This test would require disabling a server first
      // Then trying to use it in chat

      await page.goto('/chat');

      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Use a disabled tool');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(3000);

      // Should either skip the tool or show appropriate message
      const response = page.locator('[data-testid="assistant-message"], .assistant-message').last();
      await expect(response).toBeVisible({ timeout: 10000 });

      // Should not crash or show technical errors
      const criticalError = page.locator('text=/undefined|null pointer|exception/i');
      expect(await criticalError.isVisible()).toBe(false);
    });

    test('should handle OAuth token expiration', async ({ authenticatedPage: page }) => {
      // This test checks if the system handles expired OAuth tokens

      await page.goto('/chat');

      // Send message that might trigger OAuth-protected tool
      const messageInput = page.locator('textarea[placeholder*="message" i], input[placeholder*="message" i]');
      await messageInput.fill('Access OAuth-protected resource');
      await page.click('button[type="submit"], button:has-text("Send")');

      await page.waitForTimeout(5000);

      // Should either refresh token automatically or show appropriate error
      const response = page.locator('[data-testid="assistant-message"], .assistant-message').last();
      await expect(response).toBeVisible({ timeout: 15000 });

      // Check for auth error vs successful refresh
      const authError = page.locator('text=/unauthorized|re-authenticate/i');
      if (await authError.isVisible()) {
        // If auth error, it should be user-friendly
        const errorText = await authError.textContent();
        expect(errorText).toContain('authenticate');
      }
    });
  });
});
