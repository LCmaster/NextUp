// API base URL — reads from environment or defaults to localhost
const API_BASE = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const opts: RequestInit = {
		method,
		credentials: 'include', // Always send the HttpOnly session cookie
		headers: { 'Content-Type': 'application/json' }
	};
	if (body) {
		opts.body = JSON.stringify(body);
	}
	const res = await fetch(`${API_BASE}${path}`, opts);
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error || res.statusText);
	}
	if (res.status === 204) return undefined as T;
	return res.json();
}

// --- Types ---

export interface User {
	id: string;
	first_name: string;
	last_name: string;
	email: string;
	github_link: string | null;
	created_at: string;
	updated_at: string;
}

export interface Project {
	id: string;
	name: string;
	description: string | null;
	owner_id: string;
	created_at: string;
	updated_at: string;
}

export interface Ticket {
	id: string;
	project_id: string;
	title: string;
	description: string | null;
	status: string;
	priority: string;
	assignee_id: string | null;
	parent_id: string | null;
	created_at: string;
	updated_at: string;
}

// --- Users ---

export function getSetupStatus(): Promise<{ is_setup: boolean }> {
	return request('GET', '/api/v1/users/setup-status');
}

export function setupAccount(data: {
	first_name: string;
	last_name: string;
	email: string;
	password: string;
	github_link?: string;
}): Promise<User> {
	return request('POST', '/api/v1/users/setup', data);
}

export function login(data: { email: string; password: string }): Promise<User> {
	return request('POST', '/api/v1/users/login', data);
}

/**
 * Fetches the currently authenticated user's profile from the server.
 * The server identifies the user from the JWT session cookie — no arguments needed.
 */
export function getProfile(): Promise<User> {
	return request('GET', '/api/v1/users/me');
}

/**
 * Clears the session cookie on the server, effectively logging out the user.
 */
export function logout(): Promise<{ message: string }> {
	return request('POST', '/api/v1/users/logout');
}

// --- Projects ---

export function createProject(data: {
	name: string;
	description?: string;
	owner_id: string;
}): Promise<Project> {
	return request('POST', '/api/v1/projects', data);
}

export function listProjects(ownerId: string): Promise<Project[]> {
	return request('GET', `/api/v1/projects?owner_id=${ownerId}`);
}

export function getProject(id: string): Promise<Project> {
	return request('GET', `/api/v1/projects/${id}`);
}

export function updateProject(
	id: string,
	data: { name: string; description?: string }
): Promise<Project> {
	return request('PUT', `/api/v1/projects/${id}`, data);
}

export function deleteProject(id: string): Promise<void> {
	return request('DELETE', `/api/v1/projects/${id}`);
}

// --- Tickets ---

export function createTicket(data: {
	project_id: string;
	title: string;
	description?: string;
	status?: string;
	priority?: string;
	assignee_id?: string;
	parent_id?: string;
}): Promise<Ticket> {
	return request('POST', '/api/v1/tickets', data);
}

export function listTickets(projectId: string): Promise<Ticket[]> {
	return request('GET', `/api/v1/tickets?project_id=${projectId}`);
}

export function getTicket(id: string): Promise<Ticket> {
	return request('GET', `/api/v1/tickets/${id}`);
}

export function updateTicket(
	id: string,
	data: {
		title: string;
		description?: string;
		status: string;
		priority: string;
		assignee_id?: string;
		parent_id?: string;
	}
): Promise<Ticket> {
	return request('PUT', `/api/v1/tickets/${id}`, data);
}

export function deleteTicket(id: string): Promise<void> {
	return request('DELETE', `/api/v1/tickets/${id}`);
}

export function breakdownTicket(id: string): Promise<Ticket[]> {
	return request('POST', `/api/v1/tickets/${id}/breakdown`);
}
