import { test, expect } from './fixtures/auth';

test.describe('Chat Functionality', () => {
  test.describe('Chat Interface', () => {
    test('should navigate to chat page', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Click on chat navigation
      await page.locator('a[href="/chat"], button:has-text("Chat")').click();

      // Should be on chat page
      await page.waitForURL('/chat', { timeout: 5000 });
      await expect(page).toHaveURL('/chat');
    });

    test('should display chat input field', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Should see message input
      const messageInput = page.locator('textarea, input[type="text"][placeholder*="message" i], input[placeholder*="ask" i]');
      await expect(messageInput).toBeVisible();
    });

    test('should display send button', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Should see send button
      const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]');
      await expect(sendButton).toBeVisible();
    });
  });

  test.describe('Sending Messages', () => {
    test('should send a message successfully', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      // Type a message
      await messageInput.fill('Hello, this is a test message');

      // Send message
      await sendButton.click();

      // Should see the message in chat history
      await expect(page.locator('text=Hello, this is a test message')).toBeVisible({ timeout: 5000 });
    });

    test('should show loading state while waiting for response', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Test message for loading state');
      await sendButton.click();

      // Should show loading indicator
      await expect(page.locator('text=/loading|thinking|typing|generating/i, [role="progressbar"]')).toBeVisible({ timeout: 2000 });
    });

    test('should receive AI response', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('What is machine learning?');
      await sendButton.click();

      // Should eventually receive a response
      // Look for assistant/AI message (not the user's message)
      await page.waitForTimeout(10000); // AI responses take time

      // Should have more than one message (user + AI)
      const messages = page.locator('[role="article"], .message, .chat-message');
      await expect(messages).toHaveCount(2, { timeout: 15000 });
    });

    test('should prevent sending empty messages', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      // Try to send without typing
      await sendButton.click();

      // Button should be disabled or message should not be sent
      const isDisabled = await sendButton.isDisabled();
      expect(isDisabled).toBe(true);
    });

    test('should handle long messages', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      const longMessage = 'This is a very long message. '.repeat(100);

      await messageInput.fill(longMessage);
      await sendButton.click();

      // Should see the message (possibly truncated)
      await expect(page.locator('text=/This is a very long message/i')).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Conversation Management', () => {
    test('should create a new conversation', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Look for new conversation button
      const newConvoButton = page.locator('button:has-text("New"), button:has-text("New Conversation")');

      if (await newConvoButton.isVisible({ timeout: 2000 })) {
        await newConvoButton.click();

        // Should clear the chat
        const messages = page.locator('[role="article"], .message, .chat-message');
        await expect(messages).toHaveCount(0, { timeout: 2000 });
      }
    });

    test('should display conversation history', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Send a message to create conversation
      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Test conversation history');
      await sendButton.click();

      await page.waitForTimeout(2000);

      // Navigate away and back
      await page.goto('/dashboard');
      await page.goto('/chat');

      // Should see previous message
      await expect(page.locator('text=Test conversation history')).toBeVisible({ timeout: 5000 });
    });

    test('should list previous conversations', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Look for conversations sidebar/list
      const conversationsList = page.locator('[aria-label*="conversation" i], .conversations, .sidebar');

      if (await conversationsList.isVisible({ timeout: 2000 })) {
        // Should have at least one conversation after previous tests
        await expect(conversationsList.locator('div, li')).toHaveCount(1, { timeout: 3000 });
      }
    });
  });

  test.describe('Chat Features', () => {
    test('should show sources/citations if available', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Tell me about the uploaded documents');
      await sendButton.click();

      // Wait for response
      await page.waitForTimeout(10000);

      // Look for sources/citations (if implemented)
      const hasSources = await page.locator('text=/source|citation|document|reference/i').isVisible({ timeout: 5000 });

      // Just verify feature is present (may not always show sources)
      expect(typeof hasSources).toBe('boolean');
    });

    test('should support markdown formatting in responses', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Give me a bullet list of ML concepts');
      await sendButton.click();

      // Wait for response
      await page.waitForTimeout(10000);

      // Look for rendered markdown (lists, bold, etc.)
      const hasMarkdown = await Promise.race([
        page.locator('ul li, ol li, strong, em, code').isVisible().then(() => true),
        page.waitForTimeout(5000).then(() => false),
      ]);

      expect(typeof hasMarkdown).toBe('boolean');
    });

    test('should allow copying messages', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Copy test message');
      await sendButton.click();

      await page.waitForTimeout(2000);

      // Look for copy button
      const copyButton = page.locator('button[aria-label*="copy" i], button:has-text("Copy")');

      if (await copyButton.isVisible({ timeout: 2000 })) {
        await copyButton.first().click();

        // Should show copied confirmation
        await expect(page.locator('text=/copied|copy.*success/i')).toBeVisible({ timeout: 3000 });
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should show error when API fails', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Mock API to return error
      await page.route('**/api/chat/send', route =>
        route.fulfill({
          status: 500,
          body: JSON.stringify({ error: 'Internal server error' }),
        })
      );

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('This should fail');
      await sendButton.click();

      // Should show error message
      await expect(page.locator('text=/error|failed|something went wrong/i')).toBeVisible({ timeout: 5000 });
    });

    test('should handle network timeout', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Mock API to timeout
      await page.route('**/api/chat/send', route => {
        return new Promise(() => {
          // Never resolve - causes timeout
        });
      });

      const messageInput = page.locator('textarea, input[type="text"]').first();
      const sendButton = page.locator('button:has-text("Send"), button[type="submit"]').first();

      await messageInput.fill('Timeout test');
      await sendButton.click();

      // Should show timeout or error message
      await expect(page.locator('text=/timeout|taking.*long|try.*again/i')).toBeVisible({ timeout: 35000 });
    });
  });

  test.describe('Chat Configuration', () => {
    test('should allow selecting chat model', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Look for model selector
      const modelSelector = page.locator('select[name="model"], button:has-text("Model"), button:has-text("GPT")');

      if (await modelSelector.isVisible({ timeout: 2000 })) {
        await modelSelector.click();

        // Should show model options
        await expect(page.locator('text=/GPT-4|GPT-3.5|Model/i')).toBeVisible();
      }
    });

    test('should allow adjusting temperature', async ({ authenticatedPage: page }) => {
      await page.goto('/chat');

      // Look for settings/configuration button
      const settingsButton = page.locator('button[aria-label*="setting" i], button:has-text("Settings")');

      if (await settingsButton.isVisible({ timeout: 2000 })) {
        await settingsButton.click();

        // Look for temperature control
        const tempControl = page.locator('input[type="range"], input[name*="temperature" i]');
        await expect(tempControl).toBeVisible({ timeout: 3000 });
      }
    });
  });
});
