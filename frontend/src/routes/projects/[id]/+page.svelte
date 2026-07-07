<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { ticketsStore } from '$lib/stores/websocket';
	import { toast } from '$lib/stores/toast';
	import { getProject, listTickets } from '$lib/api';
	import type { Project } from '$lib/api';
	import KanbanBoard from '$lib/components/KanbanBoard.svelte';
	import KanbanSkeleton from '$lib/components/KanbanSkeleton.svelte';
	import NewTicketForm from '$lib/components/NewTicketForm.svelte';

	let projectId = $derived(page.params.id as string);
	let project = $state<Project | null>(null);
	let loading = $state(true);
	let tickets = $derived($ticketsStore || []);

	// New-ticket form state
	let showNewTicket = $state(false);
	let newTicketParentId = $state<string | null>(null);

	onMount(async () => {
		try {
			[project] = await Promise.all([
				getProject(projectId),
				listTickets(projectId).then((data) => ticketsStore.set(data || []))
			]);
		} catch (err) {
			console.error('Failed to load project:', err);
			toast.error('Failed to load project. Please try refreshing.');
		} finally {
			loading = false;
		}
	});

	function openNewTicket(parentId: string | null = null) {
		newTicketParentId = parentId;
		showNewTicket = true;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	function closeNewTicket() {
		showNewTicket = false;
		newTicketParentId = null;
	}

	function handleTicketCreated() {
		const label = newTicketParentId ? 'Sub-task created' : 'Ticket created';
		toast.success(label);
	}
</script>

<svelte:head>
	<title>{project?.name || 'Project'} — NextUp</title>
</svelte:head>

<div class="max-w-full">
	{#if loading}
		<!-- Header shimmer -->
		<div class="mb-6">
			<div class="h-9 w-56 animate-pulse rounded-lg bg-surface-200 dark:bg-surface-800"></div>
			<div class="mt-2 h-4 w-80 animate-pulse rounded bg-surface-100 dark:bg-surface-700/50"></div>
		</div>
		<div class="mb-4 h-9 w-32 animate-pulse rounded-lg bg-surface-200 dark:bg-surface-800"></div>
		<KanbanSkeleton />
	{:else if project}
		<!-- Header -->
		<div class="mb-6 flex items-center justify-between">
			<div>
				<h1 class="text-3xl font-bold tracking-tight text-surface-900 dark:text-white">
					{project.name}
				</h1>
				{#if project.description}
					<p class="mt-1 text-surface-700/60 dark:text-surface-200/40">{project.description}</p>
				{/if}
			</div>
		</div>

		<!-- Actions -->
		<div class="mb-4 flex gap-3">
			<button
				onclick={() => openNewTicket(null)}
				class="flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary-600"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				New Ticket
			</button>
			<a
				href="/projects/{projectId}/members"
				class="flex items-center gap-2 rounded-lg border border-surface-200 bg-white px-4 py-2 text-sm font-medium text-surface-700 shadow-sm transition-colors hover:bg-surface-50 dark:border-surface-700 dark:bg-surface-800 dark:text-surface-200 dark:hover:bg-surface-700/50"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
				</svg>
				Members
			</a>
		</div>

		{#if showNewTicket}
			<NewTicketForm
				{projectId}
				parentId={newTicketParentId}
				oncancel={closeNewTicket}
				oncreated={handleTicketCreated}
			/>
		{/if}

		<!-- Kanban board -->
		<KanbanBoard
			{tickets}
			onrequestsubtask={(parentId) => openNewTicket(parentId)}
		/>
	{:else}
		<div class="flex min-h-[50vh] items-center justify-center">
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">
				Project not found.
			</p>
		</div>
	{/if}
</div>
