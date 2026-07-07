import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark';

function createThemeStore() {
	const { subscribe, set, update } = writable<Theme>('light');

	if (browser) {
		const storedTheme = localStorage.getItem('theme') as Theme | null;
		if (storedTheme) {
			set(storedTheme);
			if (storedTheme === 'dark') document.documentElement.classList.add('dark');
		} else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
			set('dark');
			document.documentElement.classList.add('dark');
		}
	}

	return {
		subscribe,
		toggle: () => {
			update((currentTheme) => {
				const newTheme = currentTheme === 'light' ? 'dark' : 'light';
				if (browser) {
					localStorage.setItem('theme', newTheme);
					if (newTheme === 'dark') {
						document.documentElement.classList.add('dark');
					} else {
						document.documentElement.classList.remove('dark');
					}
				}
				return newTheme;
			});
		},
		set: (theme: Theme) => {
			if (browser) {
				localStorage.setItem('theme', theme);
				if (theme === 'dark') {
					document.documentElement.classList.add('dark');
				} else {
					document.documentElement.classList.remove('dark');
				}
			}
			set(theme);
		}
	};
}

export const themeStore = createThemeStore();
