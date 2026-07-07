import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { applyWsEvent, ticketsStore, projectsStore } from './websocket';
import type { Ticket, Project } from '$lib/api';

describe('websocket events', () => {
	beforeEach(() => {
		ticketsStore.set([]);
		projectsStore.set([]);
	});

	it('applyWsEvent ticket.created prepends the new ticket', () => {
		const ticket: Ticket = {
			id: '1', project_id: 'p1', title: 'New', status: 'todo', priority: 'medium',
			created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null, creator_id: 'u1'
		};
		
		applyWsEvent({ type: 'ticket.created', payload: ticket });
		
		const current = get(ticketsStore);
		expect(current).toHaveLength(1);
		expect(current[0].title).toBe('New');
	});

	it('applyWsEvent ticket.updated updates in place', () => {
		const ticket1: Ticket = {
			id: '1', project_id: 'p1', title: 'Old', status: 'todo', priority: 'medium',
			created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null, creator_id: 'u1'
		};
		ticketsStore.set([ticket1]);

		const updatedTicket = { ...ticket1, title: 'Updated' };
		applyWsEvent({ type: 'ticket.updated', payload: updatedTicket });

		const current = get(ticketsStore);
		expect(current).toHaveLength(1);
		expect(current[0].title).toBe('Updated');
	});

	it('applyWsEvent ticket.deleted removes by id', () => {
		const ticket1: Ticket = {
			id: '1', project_id: 'p1', title: 'Old', status: 'todo', priority: 'medium',
			created_at: '', updated_at: '', parent_id: null, assignee_id: null, description: null, creator_id: 'u1'
		};
		ticketsStore.set([ticket1]);

		applyWsEvent({ type: 'ticket.deleted', payload: { id: '1' } });

		expect(get(ticketsStore)).toHaveLength(0);
	});

	it('applyWsEvent project.created prepends the new project', () => {
		const project: Project = {
			id: 'p1', name: 'New Project', owner_id: 'u1',
			created_at: '', updated_at: '', description: null
		};
		
		applyWsEvent({ type: 'project.created', payload: project });
		
		const current = get(projectsStore);
		expect(current).toHaveLength(1);
		expect(current[0].name).toBe('New Project');
	});

	it('applyWsEvent unknown event is a no-op', () => {
		applyWsEvent({ type: 'unknown.event', payload: { foo: 'bar' } });
		
		expect(get(ticketsStore)).toHaveLength(0);
		expect(get(projectsStore)).toHaveLength(0);
	});
});
