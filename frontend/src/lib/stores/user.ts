import { writable } from 'svelte/store';
import { goto } from '$app/navigation';
import type { User } from '$lib/api';
import { getProfile, logout as apiLogout } from '$lib/api';

function createUserStore() {
	const { subscribe, set } = writable<User | null | undefined>(undefined);

	return {
		subscribe,
		set,
		login(user: User) {
			set(user);
		},
		async logout() {
			try {
				await apiLogout();
			} catch {
				// Best-effort — clear local state regardless of server response
			}
			set(null);
			goto('/login');
		},
		/**
		 * Called on app mount. Attempts to restore the session by calling /me.
		 * If the HttpOnly JWT cookie is still valid the server returns the user.
		 * On any error (401, network) the store remains null and we do nothing —
		 * the route guards will redirect to /login as needed.
		 */
		async restore() {
			try {
				const user = await getProfile();
				set(user);
			} catch {
				set(null);
			}
		}
	};
}

export const userStore = createUserStore();
