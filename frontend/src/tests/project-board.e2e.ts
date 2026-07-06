import { test, expect } from '@playwright/test';

test.describe('Project Kanban Board', () => {
	let projectId: string;

	test.beforeAll(async ({ request }) => {
		// Because we're using storageState, we're already authenticated for API calls.
		// We can get the user ID by checking setup status (which isn't authenticated)
		// but since we need a project, let's create a test project via API to ensure a clean state.
		
		// 1. Get user by calling setup and trying to list projects
		// Unfortunately, we need the owner_id to create a project via API.
		// But in our E2E environment we don't know the ID of the user created by setup.
		// Instead, we can create the project via the UI in the beforeAll step, using a dedicated page context.
	});

	test.beforeEach(async ({ page }) => {
		// Go to dashboard and create a fresh project for this test run
		await page.goto('/dashboard');
		await page.click('button:has-text("New Project")');
		await page.fill('input[placeholder="Project name"]', 'E2E Test Project');
		await page.click('button:has-text("Create Project")');
		
		// The dashboard updates automatically and shows the project card
		// We click it to navigate to the board
		const projectCard = page.locator('h3:has-text("E2E Test Project")').first();
		await projectCard.click();

		// Ensure we are on the project page
		await expect(page).toHaveURL(/\/projects\/.+/);
		// Get project ID from URL
		const url = page.url();
		projectId = url.split('/').pop() || '';
	});

	test.afterEach(async ({ request }) => {
		// Cleanup the project via API using the projectId
		if (projectId) {
			await request.delete(`/api/v1/projects/${projectId}`);
		}
	});

	test('should create a new ticket in the To Do column', async ({ page }) => {
		// Click New Ticket
		await page.click('button:has-text("New Ticket")');

		// Fill the form
		await page.fill('input[placeholder="Ticket title"]', 'My shiny new E2E ticket');
		await page.click('button:has-text("Create Ticket")');

		// A toast should appear
		await expect(page.locator('text=Ticket created')).toBeVisible();

		// The ticket should be in the To Do column (first column)
		const todoColumn = page.locator('.grid > div').nth(0);
		await expect(todoColumn).toContainText('My shiny new E2E ticket');
	});

	test('should move a ticket through the columns', async ({ page }) => {
		// First create a ticket
		await page.click('button:has-text("New Ticket")');
		await page.fill('input[placeholder="Ticket title"]', 'Moving ticket');
		await page.click('button:has-text("Create Ticket")');
		
		// Wait for it to appear
		const todoColumn = page.locator('.grid > div').nth(0);
		await expect(todoColumn).toContainText('Moving ticket');

		// Hover over the ticket to reveal buttons
		const ticketCard = page.locator('div:has-text("Moving ticket")').last();
		await ticketCard.hover();

		// Click the 'Move right' button (the one pointing right)
		// We look for the SVG path that looks like a right arrow.
		// Based on our implementation, there are move left and right buttons.
		const moveRightBtn = ticketCard.locator('button[aria-label="Move right"]');
		await moveRightBtn.click();

		// It should now be in the "In Progress" column (index 1)
		const inProgressColumn = page.locator('.grid > div').nth(1);
		await expect(inProgressColumn).toContainText('Moving ticket');
		
		// Move it to "Done" (index 2)
		const ticketInProg = page.locator('div:has-text("Moving ticket")').last();
		await ticketInProg.hover();
		const moveRightBtn2 = ticketInProg.locator('button[aria-label="Move right"]');
		await moveRightBtn2.click();

		const doneColumn = page.locator('.grid > div').nth(2);
		await expect(doneColumn).toContainText('Moving ticket');
	});
});
