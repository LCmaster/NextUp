import { writable } from 'svelte/store';
import type { Ticket, Todo, Project } from '$lib/api';

export interface WsEvent {
	type: string;
	payload: unknown;
}

const WS_BASE = (import.meta.env.PUBLIC_API_URL || 'http://localhost:8080').replace(
	/^http/,
	'ws'
);

function createWebSocketStore() {
	let socket: WebSocket | null = null;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	const RECONNECT_DELAY = 3000;

	const { subscribe, set } = writable<WsEvent | null>(null);

	function connect() {
		if (socket?.readyState === WebSocket.OPEN) return;

		socket = new WebSocket(`${WS_BASE}/ws`);

		socket.onopen = () => {
			console.log('[WS] Connected');
		};

		socket.onmessage = (event) => {
			try {
				const data: WsEvent = JSON.parse(event.data);
				set(data);
			} catch (err) {
				console.error('[WS] Failed to parse message:', err);
			}
		};

		socket.onclose = () => {
			console.log('[WS] Disconnected, reconnecting...');
			scheduleReconnect();
		};

		socket.onerror = (err) => {
			console.error('[WS] Error:', err);
			socket?.close();
		};
	}

	function scheduleReconnect() {
		if (reconnectTimer) clearTimeout(reconnectTimer);
		reconnectTimer = setTimeout(() => connect(), RECONNECT_DELAY);
	}

	function disconnect() {
		if (reconnectTimer) clearTimeout(reconnectTimer);
		socket?.close();
		socket = null;
	}

	return {
		subscribe,
		connect,
		disconnect
	};
}

export const wsStore = createWebSocketStore();

// --- Reactive entity stores updated by WebSocket events ---

export const ticketsStore = writable<Ticket[]>([]);
export const todosStore = writable<Todo[]>([]);
export const projectsStore = writable<Project[]>([]);

// Apply a WebSocket event to the local stores
export function applyWsEvent(event: WsEvent) {
	const { type, payload } = event;

	switch (type) {
		// Tickets
		case 'ticket.created':
			ticketsStore.update((tickets) => [payload as Ticket, ...tickets]);
			break;
		case 'ticket.updated':
			ticketsStore.update((tickets) =>
				tickets.map((t) => (t.id === (payload as Ticket).id ? (payload as Ticket) : t))
			);
			break;
		case 'ticket.deleted':
			ticketsStore.update((tickets) =>
				tickets.filter((t) => t.id !== (payload as { id: string }).id)
			);
			break;

		// Todos
		case 'todo.created':
			todosStore.update((todos) => [payload as Todo, ...todos]);
			break;
		case 'todo.updated':
			todosStore.update((todos) =>
				todos.map((t) => (t.id === (payload as Todo).id ? (payload as Todo) : t))
			);
			break;
		case 'todo.deleted':
			todosStore.update((todos) =>
				todos.filter((t) => t.id !== (payload as { id: string }).id)
			);
			break;

		// Projects
		case 'project.created':
			projectsStore.update((projects) => [payload as Project, ...projects]);
			break;
		case 'project.updated':
			projectsStore.update((projects) =>
				projects.map((p) =>
					p.id === (payload as Project).id ? (payload as Project) : p
				)
			);
			break;
		case 'project.deleted':
			projectsStore.update((projects) =>
				projects.filter((p) => p.id !== (payload as { id: string }).id)
			);
			break;
	}
}
