import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './src/tests',
	testMatch: '**/*.e2e.ts',
	fullyParallel: false,
	retries: process.env.CI ? 1 : 0,
	workers: 1, // To avoid state conflict if multiple tests run against the same DB
	reporter: 'html',
	use: {
		baseURL: 'http://localhost:80',
		trace: 'on-first-retry'
	},
	projects: [
		{ name: 'setup', testMatch: /.*\.setup\.ts/ },
		{
			name: 'chromium',
			use: {
				...devices['Desktop Chrome'],
				// Use the signed-in state for all tests in this project
				storageState: 'playwright/.auth/user.json'
			},
			dependencies: ['setup']
		}
	]
});
