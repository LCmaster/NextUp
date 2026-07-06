import { writable } from 'svelte/store';
import type { User } from '$lib/api';

function createUserStore() {
	const { subscribe, set } = writable<User | null>(null);

	return {
		subscribe,
		set,
		login(user: User) {
			set(user);
			if (typeof localStorage !== 'undefined') {
				localStorage.setItem('nextup_user', JSON.stringify(user));
			}
		},
		logout() {
			set(null);
			if (typeof localStorage !== 'undefined') {
				localStorage.removeItem('nextup_user');
			}
		},
		restore() {
			if (typeof localStorage !== 'undefined') {
				const stored = localStorage.getItem('nextup_user');
				if (stored) {
					try {
						set(JSON.parse(stored));
					} catch {
						localStorage.removeItem('nextup_user');
					}
				}
			}
		}
	};
}

export const userStore = createUserStore();
