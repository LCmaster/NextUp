import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
	id: string;
	message: string;
	type: ToastType;
}

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function add(message: string, type: ToastType = 'info', duration = 3500) {
		const id = crypto.randomUUID();
		update((toasts) => [...toasts, { id, message, type }]);
		setTimeout(() => remove(id), duration);
	}

	function remove(id: string) {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}

	return {
		subscribe,
		success: (msg: string) => add(msg, 'success'),
		error: (msg: string) => add(msg, 'error'),
		info: (msg: string) => add(msg, 'info'),
		remove
	};
}

export const toast = createToastStore();
