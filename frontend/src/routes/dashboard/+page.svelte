<script lang="ts">
	import { onMount } from 'svelte';
	import { userStore } from '$lib/stores/user';
	import { projectsStore } from '$lib/stores/websocket';
	import { listProjects, createProject, deleteProject } from '$lib/api';
	import type { Project } from '$lib/api';

	let user = $derived($userStore);
	let projects: Project[] = $derived($projectsStore || []);

	let showNewProject = $state(false);
	let newProjectName = $state('');
	let newProjectDesc = $state('');
	let creating = $state(false);
	let loading = $state(true);

	$effect(() => {
		if (user) {
			listProjects()
				.then((data) => {
					projectsStore.set(data || []);
				})
				.catch((err) => console.error('Failed to load projects:', err))
				.finally(() => {
					loading = false;
				});
		}
	});

	async function handleCreateProject(e: SubmitEvent) {
		e.preventDefault();
		if (!user || !newProjectName.trim()) return;

		creating = true;
		try {
			await createProject({
				name: newProjectName.trim(),
				description: newProjectDesc.trim() || undefined
			});
			newProjectName = '';
			newProjectDesc = '';
			showNewProject = false;
		} catch (err) {
			console.error('Failed to create project:', err);
		} finally {
			creating = false;
		}
	}

	async function handleDeleteProject(id: string) {
		try {
			await deleteProject(id);
		} catch (err) {
			console.error('Failed to delete project:', err);
		}
	}
</script>

<svelte:head>
	<title>Dashboard — NextUp</title>
</svelte:head>

<div class="max-w-5xl">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold tracking-tight text-surface-900 dark:text-white">
			Welcome back{user ? `, ${user.first_name}` : ''}
		</h1>
		<p class="mt-1 text-surface-700/60 dark:text-surface-200/40">
			Here's an overview of your projects.
		</p>
	</div>

	<!-- Stats row -->
	<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
		<div
			class="rounded-xl border border-surface-200/60 bg-white p-5 shadow-sm dark:border-surface-800 dark:bg-surface-900"
		>
			<p class="text-sm font-medium text-surface-700/60 dark:text-surface-200/40">Total Projects</p>
			<p class="mt-1 text-2xl font-bold text-surface-900 dark:text-white">{projects.length}</p>
		</div>
		<div
			class="rounded-xl border border-surface-200/60 bg-white p-5 shadow-sm dark:border-surface-800 dark:bg-surface-900"
		>
			<p class="text-sm font-medium text-surface-700/60 dark:text-surface-200/40">Active</p>
			<p class="mt-1 text-2xl font-bold text-primary-600 dark:text-primary-400">
				{projects.length}
			</p>
		</div>
		<div
			class="rounded-xl border border-surface-200/60 bg-white p-5 shadow-sm dark:border-surface-800 dark:bg-surface-900"
		>
			<p class="text-sm font-medium text-surface-700/60 dark:text-surface-200/40">Completed</p>
			<p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">0</p>
		</div>
	</div>

	<!-- Projects section -->
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-lg font-semibold text-surface-900 dark:text-white">Projects</h2>
		<button
			onclick={() => (showNewProject = !showNewProject)}
			class="flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary-600"
		>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
			</svg>
			New Project
		</button>
	</div>

	<!-- New project form -->
	{#if showNewProject}
		<div
			class="mb-6 rounded-xl border border-primary-200 bg-primary-50/50 p-6 dark:border-primary-800/30 dark:bg-primary-900/10"
		>
			<form onsubmit={handleCreateProject} class="flex flex-col gap-4">
				<input
					type="text"
					bind:value={newProjectName}
					placeholder="Project name"
					required
					class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
				/>
				<textarea
					bind:value={newProjectDesc}
					placeholder="Description (optional)"
					rows="2"
					class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
				></textarea>
				<div class="flex gap-3">
					<button
						type="submit"
						disabled={creating}
						class="rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
					>
						{creating ? 'Creating...' : 'Create Project'}
					</button>
					<button
						type="button"
						onclick={() => (showNewProject = false)}
						class="rounded-lg px-4 py-2 text-sm font-medium text-surface-700 transition-colors hover:bg-surface-200 dark:text-surface-200 dark:hover:bg-surface-800"
					>
						Cancel
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Project cards -->
	{#if loading}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each [1, 2, 3] as _}
				<div
					class="rounded-xl border border-surface-200/60 bg-white p-6 shadow-sm dark:border-surface-800 dark:bg-surface-900"
				>
					<div class="mb-3 flex items-start justify-between">
						<div
							class="h-10 w-10 animate-pulse rounded-lg bg-surface-200 dark:bg-surface-800"
						></div>
					</div>
					<div
						class="mt-4 h-5 w-1/2 animate-pulse rounded bg-surface-200 dark:bg-surface-800"
					></div>
					<div
						class="mt-2 h-4 w-3/4 animate-pulse rounded bg-surface-200 dark:bg-surface-800"
					></div>
					<div
						class="mt-4 h-3 w-1/4 animate-pulse rounded bg-surface-200 dark:bg-surface-800"
					></div>
				</div>
			{/each}
		</div>
	{:else if projects.length === 0}
		<div
			class="rounded-xl border border-dashed border-surface-200 py-16 text-center dark:border-surface-800"
		>
			<svg
				class="mx-auto h-12 w-12 text-surface-700/30 dark:text-surface-200/20"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				stroke-width="1"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z"
				/>
			</svg>
			<p class="mt-4 text-sm text-surface-700/60 dark:text-surface-200/40">
				No projects yet. Create your first one!
			</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each projects as project (project.id)}
				<a
					href="/projects/{project.id}"
					class="group rounded-xl border border-surface-200/60 bg-white p-6 shadow-sm transition-all duration-200 hover:border-primary-300 hover:shadow-md dark:border-surface-800 dark:bg-surface-900 dark:hover:border-primary-700"
				>
					<div class="mb-3 flex items-start justify-between">
						<div
							class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-primary-400 to-primary-600 text-sm font-bold text-white"
						>
							{project.name[0]?.toUpperCase() || 'P'}
						</div>
						<button
							onclick={(e) => {
								e.preventDefault();
								e.stopPropagation();
								handleDeleteProject(project.id);
							}}
							class="rounded-md p-1.5 text-surface-700/30 opacity-0 transition-all hover:bg-red-50 hover:text-red-500 group-hover:opacity-100 dark:text-surface-200/20 dark:hover:bg-red-900/20 dark:hover:text-red-400"
							aria-label="Delete project"
						>
							<svg
								class="h-4 w-4"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								stroke-width="1.5"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
								/>
							</svg>
						</button>
					</div>
					<h3 class="text-base font-semibold text-surface-900 dark:text-white">{project.name}</h3>
					{#if project.description}
						<p class="mt-1 line-clamp-2 text-sm text-surface-700/60 dark:text-surface-200/40">
							{project.description}
						</p>
					{/if}
					<p class="mt-3 text-xs text-surface-700/40 dark:text-surface-200/30">
						Created {new Date(project.created_at).toLocaleDateString()}
					</p>
				</a>
			{/each}
		</div>
	{/if}
</div>
