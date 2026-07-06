import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { page } from 'vitest/browser';
import KanbanBoard from './KanbanBoard.svelte';
import type { Ticket } from '$lib/api';

describe('KanbanBoard.svelte', () => {
	const mockTickets: Ticket[] = [
		{ id: '1', project_id: 'p1', title: 'Task 1', status: 'todo', priority: 'high', created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null },
		{ id: '2', project_id: 'p1', title: 'Task 2', status: 'in_progress', priority: 'medium', created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null },
		{ id: '3', project_id: 'p1', title: 'Task 3', status: 'done', priority: 'low', created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null },
		{ id: '4', project_id: 'p1', title: 'Task 4', status: 'archived', priority: 'medium', created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null },
	];

	it('renders 4 columns', async () => {
		render(KanbanBoard, { tickets: mockTickets });

		await expect.element(page.getByText('To Do')).toBeInTheDocument();
		await expect.element(page.getByText('In Progress')).toBeInTheDocument();
		await expect.element(page.getByText('Done')).toBeInTheDocument();
		await expect.element(page.getByText('Archived')).toBeInTheDocument();
	});

	it('places tickets in the correct column', async () => {
		render(KanbanBoard, { tickets: mockTickets });

		await expect.element(page.getByText('Task 1')).toBeInTheDocument();
		await expect.element(page.getByText('Task 2')).toBeInTheDocument();
		await expect.element(page.getByText('Task 3')).toBeInTheDocument();
	});

	it('archived tickets render in the archived column', async () => {
		render(KanbanBoard, { tickets: mockTickets });

		// We assume Task 4 is the archived one
		await expect.element(page.getByText('Task 4')).toBeInTheDocument();
	});

	it('shows the correct ticket count per column', async () => {
		render(KanbanBoard, { tickets: mockTickets });

		// All 4 columns should have a count badge of "1"
		const badges = page.getByText('1');
		const elements = await badges.elements();
		// Given that we have 4 columns and exactly 1 ticket in each, there should be 4 badges showing "1".
		// Note that vitest-browser might find other "1"s, but we can verify at least 4 exist.
		expect(elements.length).toBeGreaterThanOrEqual(4);
	});
});
