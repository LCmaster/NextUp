<script lang="ts">
	import { page } from '$app/state';
	import { userStore } from '$lib/stores/user';
	import { projectsStore } from '$lib/stores/websocket';
	import type { Project } from '$lib/api';

	let user = $derived($userStore);
	let projects: Project[] = $derived($projectsStore);
	let currentPath = $derived(page.url.pathname);

	function isActive(path: string): boolean {
		return currentPath === path || currentPath.startsWith(path + '/');
	}

	async function handleLogout() {
		await userStore.logout();
	}
</script>

<aside
	class="fixed left-0 top-0 flex h-screen w-64 flex-col border-r border-surface-200 bg-surface-50 dark:border-surface-800 dark:bg-surface-950"
>
	<!-- Logo -->
	<div class="flex items-center gap-3 border-b border-surface-200 px-6 py-5 dark:border-surface-800">
		<div
			class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 text-sm font-bold text-white shadow-md"
		>
			N
		</div>
		<span class="text-lg font-bold tracking-tight text-surface-900 dark:text-white">NextUp</span>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 overflow-y-auto px-3 py-4">
		<div class="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-surface-700/60 dark:text-surface-200/40">
			Menu
		</div>

		<a
			href="/dashboard"
			class="mb-1 flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-150
				{isActive('/dashboard')
				? 'bg-primary-500/10 text-primary-600 dark:text-primary-400'
				: 'text-surface-700 hover:bg-surface-200/60 dark:text-surface-200 dark:hover:bg-surface-800/60'}"
		>
			<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
				<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25a2.25 2.25 0 0 1-2.25-2.25v-2.25Z" />
			</svg>
			Dashboard
		</a>

		{#if projects.length > 0}
			<div class="mb-2 mt-6 px-3 text-xs font-semibold uppercase tracking-wider text-surface-700/60 dark:text-surface-200/40">
				Projects
			</div>
			{#each projects as project (project.id)}
				<a
					href="/projects/{project.id}"
					class="mb-1 flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-150
						{isActive(`/projects/${project.id}`)
						? 'bg-primary-500/10 text-primary-600 dark:text-primary-400'
						: 'text-surface-700 hover:bg-surface-200/60 dark:text-surface-200 dark:hover:bg-surface-800/60'}"
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" />
					</svg>
					{project.name}
				</a>
			{/each}
		{/if}
	</nav>

	<!-- User info footer with logout -->
	{#if user}
		<div class="border-t border-surface-200 px-4 py-4 dark:border-surface-800">
			<div class="flex items-center gap-3">
				<div
					class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-sm font-semibold text-white"
				>
					{user.first_name[0]}{user.last_name[0]}
				</div>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-medium text-surface-900 dark:text-white">
						{user.first_name} {user.last_name}
					</p>
					<p class="truncate text-xs text-surface-700/60 dark:text-surface-200/50">{user.email}</p>
				</div>
				<button
					id="logout-button"
					onclick={handleLogout}
					title="Sign out"
					class="flex-shrink-0 rounded-md p-1.5 text-surface-700/50 transition-colors hover:bg-red-50 hover:text-red-500 dark:text-surface-200/40 dark:hover:bg-red-900/20 dark:hover:text-red-400"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
					</svg>
				</button>
			</div>
		</div>
	{/if}
</aside>
