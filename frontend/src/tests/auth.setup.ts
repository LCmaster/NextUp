import { test as setup, expect } from '@playwright/test';
import * as path from 'path';

const authFile = path.join(__dirname, '../../playwright/.auth/user.json');

setup('authenticate', async ({ page, request }) => {
	// Check if setup is needed by hitting the API
	const res = await request.get('/api/v1/setup');
	const { is_setup } = await res.json();

	if (!is_setup) {
		// Need to create the first user
		await page.goto('/setup');
		await page.fill('input[type="text"]', 'E2E');
		await page.fill('input[placeholder="Last name"]', 'Tester');
		await page.fill('input[type="email"]', 'test@example.com');
		await page.fill('input[type="password"]', 'password123');
		await page.click('button:has-text("Complete Setup")');
	} else {
		// Just log in
		await page.goto('/login');
		await page.fill('input[type="email"]', 'test@example.com');
		await page.fill('input[type="password"]', 'password123');
		await page.click('button:has-text("Sign in")');
	}

	// Wait for dashboard to load indicating successful login
	await expect(page.locator('h1')).toContainText('Welcome back');

	// Save the storage state (which captures localStorage because of NextUp's architecture)
	await page.context().storageState({ path: authFile });
});
