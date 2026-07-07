import { writable } from 'svelte/store';
import type { Ticket, Project, ProjectMember } from '$lib/api';

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
export const projectsStore = writable<Project[]>([]);
export const projectMembersStore = writable<Record<string, ProjectMember[]>>({});

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

		// Members
		case 'project.member.added': {
			const m = payload as ProjectMember;
			projectMembersStore.update(s => {
				const members = s[m.project_id] || [];
				return { ...s, [m.project_id]: [...members, m] };
			});
			break;
		}
		case 'project.member.updated': {
			const m = payload as ProjectMember;
			projectMembersStore.update(s => {
				const members = s[m.project_id] || [];
				return {
					...s,
					[m.project_id]: members.map(x => x.user_id === m.user_id ? m : x)
				};
			});
			break;
		}
		case 'project.member.removed': {
			const removed = payload as { project_id: string; user_id: string };
			projectMembersStore.update(s => {
				const members = s[removed.project_id] || [];
				return {
					...s,
					[removed.project_id]: members.filter(x => x.user_id !== removed.user_id)
				};
			});
			break;
		}
	}
}
