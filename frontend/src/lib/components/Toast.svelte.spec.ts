import { describe, it, expect, beforeEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { page, userEvent } from 'vitest/browser';
import Toast from './Toast.svelte';
import { toast } from '../stores/toast';
import { get } from 'svelte/store';

describe('Toast.svelte', () => {
	beforeEach(() => {
		// Clear toasts before each test
		const currentToasts = get(toast);
		currentToasts.forEach((t) => toast.remove(t.id));
	});

	it('renders a success toast', async () => {
		toast.success('Successfully saved');
		render(Toast);

		const element = page.getByText('Successfully saved');
		await expect.element(element).toBeInTheDocument();
		// Vitest-browser can assert classes
		await expect.element(page.getByRole('button')).toBeInTheDocument();
	});

	it('renders an error toast', async () => {
		toast.error('Something went wrong');
		render(Toast);

		const element = page.getByText('Something went wrong');
		await expect.element(element).toBeInTheDocument();
	});

	it('dismiss button removes the toast', async () => {
		toast.info('Close me');
		render(Toast);

		const element = page.getByText('Close me');
		await expect.element(element).toBeInTheDocument();

		const closeBtn = page.getByRole('button');
		await userEvent.click(closeBtn);

		await expect.element(element).not.toBeInTheDocument();
	});
});
