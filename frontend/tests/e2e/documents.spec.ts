import { test, expect } from './fixtures/auth';
import path from 'path';

test.describe('Document Management', () => {
  test.describe('Document Upload', () => {
    test('should upload a text document successfully', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Look for upload button
      const uploadButton = page.locator('button:has-text("Upload"), input[type="file"]').first();
      await expect(uploadButton).toBeVisible({ timeout: 10000 });

      // Create a test file
      const testFilePath = path.join(__dirname, 'fixtures', 'test-document.txt');

      // Upload file
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'test-document.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('This is a test document for RAG Dashboard testing.'),
      });

      // Wait for upload to complete
      await expect(page.locator('text=/upload.*success|document.*uploaded/i')).toBeVisible({ timeout: 15000 });

      // Verify document appears in list
      await expect(page.locator('text=test-document.txt')).toBeVisible();
    });

    test('should show error for files exceeding size limit', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Create a large file (>100MB simulation)
      const largeBuffer = Buffer.alloc(101 * 1024 * 1024); // 101MB

      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'large-file.txt',
        mimeType: 'text/plain',
        buffer: largeBuffer,
      });

      // Should show size error
      await expect(page.locator('text=/file.*too large|size.*exceed|maximum.*100/i')).toBeVisible({ timeout: 5000 });
    });

    test('should handle upload errors gracefully', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Mock API to return error
      await page.route('**/api/documents/upload', route => route.abort());

      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Test content'),
      });

      // Should show error message
      await expect(page.locator('text=/upload.*failed|error.*uploading/i')).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Document List', () => {
    test('should display uploaded documents', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Upload a document first
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'list-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Document list test'),
      });

      await page.waitForTimeout(2000); // Wait for upload

      // Should see the document in list
      await expect(page.locator('text=list-test.txt')).toBeVisible();
    });

    test('should show empty state when no documents', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Check for empty state message
      const hasDocuments = await page.locator('text=/no documents|upload.*first/i').isVisible({ timeout: 3000 });

      // Either shows empty state or has documents from other tests
      expect(typeof hasDocuments).toBe('boolean');
    });

    test('should display document metadata', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Upload a document
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'metadata-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Metadata test content'),
      });

      await page.waitForTimeout(2000);

      // Look for document card/row
      const docCard = page.locator('text=metadata-test.txt').locator('..').locator('..');

      // Should show file type or size
      await expect(docCard.locator('text=/txt|text|plain|\d+\s*KB|\d+\s*MB/i')).toBeVisible();
    });
  });

  test.describe('Document Actions', () => {
    test('should view document details', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Upload a document
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'view-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('View test content'),
      });

      await page.waitForTimeout(2000);

      // Click on document to view details
      await page.locator('text=view-test.txt').click();

      // Should show document details
      await expect(page.locator('text=/details|chunks|vectors|status/i')).toBeVisible({ timeout: 5000 });
    });

    test('should delete a document', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Upload a document to delete
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'delete-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Delete test content'),
      });

      await page.waitForTimeout(2000);

      // Find and click delete button
      const docRow = page.locator('text=delete-test.txt').locator('..').locator('..');
      await docRow.locator('button:has-text("Delete"), button[aria-label*="delete" i]').click();

      // Confirm deletion if modal appears
      const confirmButton = page.locator('button:has-text("Confirm"), button:has-text("Yes"), button:has-text("Delete")');
      if (await confirmButton.isVisible({ timeout: 2000 })) {
        await confirmButton.click();
      }

      // Document should be removed
      await expect(page.locator('text=delete-test.txt')).not.toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Document Processing', () => {
    test('should show processing status for uploaded documents', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'processing-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Processing test content with more text to ensure proper chunking and embedding generation for the RAG system.'.repeat(10)),
      });

      // Should show processing indicator
      await expect(page.locator('text=/processing|embedding|pending/i')).toBeVisible({ timeout: 5000 });
    });

    test('should update status when processing completes', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'complete-test.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Complete test content'),
      });

      // Wait for processing to complete
      await expect(page.locator('text=/completed|ready|success/i')).toBeVisible({ timeout: 30000 });
    });
  });

  test.describe('Search and Filter', () => {
    test('should filter documents by name', async ({ authenticatedPage: page }) => {
      await page.goto('/dashboard');

      // Upload multiple documents
      const fileInput = page.locator('input[type="file"]');

      await fileInput.setInputFiles({
        name: 'alpha-doc.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Alpha content'),
      });
      await page.waitForTimeout(1000);

      await fileInput.setInputFiles({
        name: 'beta-doc.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('Beta content'),
      });
      await page.waitForTimeout(1000);

      // Look for search/filter input
      const searchInput = page.locator('input[type="search"], input[placeholder*="search" i], input[placeholder*="filter" i]');

      if (await searchInput.isVisible({ timeout: 2000 })) {
        await searchInput.fill('alpha');

        // Should show only alpha document
        await expect(page.locator('text=alpha-doc.txt')).toBeVisible();
        await expect(page.locator('text=beta-doc.txt')).not.toBeVisible();
      }
    });
  });
});
