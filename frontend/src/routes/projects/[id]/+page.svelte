<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { ticketsStore } from '$lib/stores/websocket';
	import { getProject, listTickets } from '$lib/api';
	import type { Project } from '$lib/api';
	import KanbanBoard from '$lib/components/KanbanBoard.svelte';
	import NewTicketForm from '$lib/components/NewTicketForm.svelte';

	let projectId = $derived(page.params.id as string);
	let project = $state<Project | null>(null);
	let tickets = $derived($ticketsStore || []);

	// New-ticket form state
	let showNewTicket = $state(false);
	let newTicketParentId = $state<string | null>(null);

	onMount(async () => {
		try {
			project = await getProject(projectId);
			const ticketData = await listTickets(projectId);
			ticketsStore.set(ticketData || []);
		} catch (err) {
			console.error('Failed to load project:', err);
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
</script>

<svelte:head>
	<title>{project?.name || 'Project'} — NextUp</title>
</svelte:head>

{#if project}
	<div class="max-w-full">
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

		<!-- New Ticket button -->
		<div class="mb-4">
			<button
				onclick={() => openNewTicket(null)}
				class="flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary-600"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				New Ticket
			</button>
		</div>

		{#if showNewTicket}
			<NewTicketForm
				{projectId}
				parentId={newTicketParentId}
				oncancel={closeNewTicket}
			/>
		{/if}

		<!-- Kanban board -->
		<KanbanBoard
			{tickets}
			onrequestsubtask={(parentId) => openNewTicket(parentId)}
		/>
	</div>
{:else}
	<div class="flex min-h-[50vh] items-center justify-center">
		<p class="animate-pulse text-sm text-surface-700/60 dark:text-surface-200/40">
			Loading project...
		</p>
	</div>
{/if}
