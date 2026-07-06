import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { page, userEvent } from 'vitest/browser';
import NewTicketForm from './NewTicketForm.svelte';
import * as api from '$lib/api';

vi.mock('$lib/api', () => ({
	createTicket: vi.fn()
}));

describe('NewTicketForm.svelte', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('renders create ticket form', async () => {
		render(NewTicketForm, { projectId: 'p1', parentId: null, oncancel: () => {} });

		await expect.element(page.getByPlaceholder('Ticket title')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Create Ticket' })).toBeInTheDocument();
	});

	it('shows "New Sub-task" heading when parentId is set', async () => {
		render(NewTicketForm, { projectId: 'p1', parentId: 'parent1', oncancel: () => {} });

		await expect.element(page.getByText('New Sub-task')).toBeInTheDocument();
	});

	it('submit calls createTicket with correct params', async () => {
		vi.mocked(api.createTicket).mockResolvedValue({
			id: '1', project_id: 'p1', title: 'My ticket', status: 'todo', priority: 'medium',
			created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null
		});

		const oncreated = vi.fn();
		const oncancel = vi.fn();
		render(NewTicketForm, { projectId: 'p1', parentId: null, oncreated, oncancel });

		const titleInput = page.getByPlaceholder('Ticket title');
		await userEvent.fill(titleInput, 'My ticket');
		
		const submitBtn = page.getByRole('button', { name: 'Create Ticket' });
		await userEvent.click(submitBtn);

		expect(api.createTicket).toHaveBeenCalledWith({
			project_id: 'p1',
			title: 'My ticket',
			description: undefined,
			priority: 'medium',
			parent_id: undefined
		});
		expect(oncreated).toHaveBeenCalled();
	});

	it('shows error message on API failure', async () => {
		vi.mocked(api.createTicket).mockRejectedValue(new Error('API failure'));

		render(NewTicketForm, { projectId: 'p1', parentId: null, oncancel: () => {} });

		const titleInput = page.getByPlaceholder('Ticket title');
		await userEvent.fill(titleInput, 'My ticket');
		
		const submitBtn = page.getByRole('button', { name: 'Create Ticket' });
		await userEvent.click(submitBtn);

		await expect.element(page.getByText('API failure')).toBeInTheDocument();
	});

	it('cancel button triggers oncancel callback', async () => {
		const oncancel = vi.fn();
		render(NewTicketForm, { projectId: 'p1', parentId: null, oncancel });

		const cancelBtn = page.getByRole('button', { name: 'Cancel' });
		await userEvent.click(cancelBtn);

		expect(oncancel).toHaveBeenCalled();
	});
});
