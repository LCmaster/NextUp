<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { ticketsStore } from '$lib/stores/websocket';
	import {
		getProject,
		listTickets,
		createTicket,
		updateTicket,
		deleteTicket,
		breakdownTicket
	} from '$lib/api';
	import type { Project, Ticket } from '$lib/api';

	const STATUSES = ['todo', 'in_progress', 'done', 'archived'] as const;
	const STATUS_LABELS: Record<string, string> = {
		todo: 'To Do',
		in_progress: 'In Progress',
		done: 'Done',
		archived: 'Archived'
	};
	const STATUS_COLORS: Record<string, string> = {
		todo: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
		in_progress: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
		done: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
		archived: 'bg-slate-200 text-slate-600 dark:bg-surface-800 dark:text-surface-400 opacity-70'
	};
	const PRIORITY_COLORS: Record<string, string> = {
		low: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
		medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
		high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
		critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
	};

	let projectId = $derived(page.params.id as string);
	let project = $state<Project | null>(null);
	let tickets: Ticket[] = $derived($ticketsStore || []);


	// Kanban columns
	let todoTickets = $derived(tickets.filter((t) => t.status === 'todo' && !t.parent_id));
	let inProgressTickets = $derived(tickets.filter((t) => t.status === 'in_progress' && !t.parent_id));
	let doneTickets = $derived(tickets.filter((t) => t.status === 'done' && !t.parent_id));
	let archivedTickets = $derived(tickets.filter((t) => t.status === 'archived' && !t.parent_id));



	// New ticket form
	let showNewTicket = $state(false);
	let newTicketParentId = $state<string | null>(null);
	let newTicketTitle = $state('');
	let newTicketDesc = $state('');
	let newTicketPriority = $state('medium');
	let creatingTicket = $state(false);
	let breakingDownId = $state<string | null>(null);

	// Edit ticket state
	let editingTicketId = $state<string | null>(null);
	let editTicketTitle = $state('');
	let editTicketDesc = $state('');
	let editTicketPriority = $state('medium');
	let isSavingEdit = $state(false);

	function startEditing(ticket: Ticket) {
		editingTicketId = ticket.id;
		editTicketTitle = ticket.title;
		editTicketDesc = ticket.description || '';
		editTicketPriority = ticket.priority;
	}

	function cancelEditing() {
		editingTicketId = null;
	}

	async function saveEdit(ticket: Ticket) {
		if (!editTicketTitle.trim()) return;
		isSavingEdit = true;
		try {
			await updateTicket(ticket.id, {
				title: editTicketTitle.trim(),
				description: editTicketDesc.trim() || undefined,
				status: ticket.status,
				priority: editTicketPriority,
				assignee_id: ticket.assignee_id || undefined,
				parent_id: ticket.parent_id || undefined
			});
			editingTicketId = null;
		} catch (err) {
			console.error('Failed to update ticket:', err);
		} finally {
			isSavingEdit = false;
		}
	}

	// Expanded tickets
	let expandedTickets = $state<string[]>([]);
	function toggleTicketExpanded(id: string) {
		if (expandedTickets.includes(id)) {
			expandedTickets = expandedTickets.filter((t) => t !== id);
		} else {
			expandedTickets = [...expandedTickets, id];
		}
	}

	onMount(async () => {
		try {
			project = await getProject(projectId);
			const ticketData = await listTickets(projectId);
			ticketsStore.set(ticketData || []);
		} catch (err) {
			console.error('Failed to load project:', err);
		}
	});

	async function handleCreateTicket(e: SubmitEvent) {
		e.preventDefault();
		if (!newTicketTitle.trim()) return;
		creatingTicket = true;
		try {
			await createTicket({
				project_id: projectId,
				title: newTicketTitle.trim(),
				description: newTicketDesc.trim() || undefined,
				priority: newTicketPriority,
				parent_id: newTicketParentId || undefined
			});
			newTicketTitle = '';
			newTicketDesc = '';
			newTicketPriority = 'medium';
			newTicketParentId = null;
			showNewTicket = false;
		} catch (err) {
			console.error('Failed to create ticket:', err);
		} finally {
			creatingTicket = false;
		}
	}

	async function handleBreakdown(ticketId: string) {
		breakingDownId = ticketId;
		try {
			await breakdownTicket(ticketId);
		} catch (err) {
			console.error('Failed to breakdown ticket:', err);
		} finally {
			breakingDownId = null;
		}
	}

	async function moveTicket(ticket: Ticket, newStatus: string) {
		try {
			await updateTicket(ticket.id, {
				title: ticket.title,
				description: ticket.description || undefined,
				status: newStatus,
				priority: ticket.priority,
				assignee_id: ticket.assignee_id || undefined
			});
		} catch (err) {
			console.error('Failed to move ticket:', err);
		}
	}

	async function handleDeleteTicket(id: string) {
		try {
			await deleteTicket(id);
		} catch (err) {
			console.error('Failed to delete ticket:', err);
		}
	}



	function getLeftStatus(status: string) {
		if (status === 'archived') return 'done';
		if (status === 'done') return 'in_progress';
		return 'todo';
	}
	
	function getRightStatus(status: string) {
		if (status === 'todo') return 'in_progress';
		if (status === 'in_progress') return 'done';
		return 'archived';
	}

	function getColumnTickets(status: string): Ticket[] {
		if (status === 'todo') return todoTickets;
		if (status === 'in_progress') return inProgressTickets;
		if (status === 'done') return doneTickets;
		return archivedTickets;
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
				<h1 class="text-3xl font-bold tracking-tight text-surface-900 dark:text-white">{project.name}</h1>
				{#if project.description}
					<p class="mt-1 text-surface-700/60 dark:text-surface-200/40">{project.description}</p>
				{/if}
			</div>
		</div>

		<!-- Board -->
			<div class="mb-4">
				<button
					onclick={() => { newTicketParentId = null; showNewTicket = !showNewTicket; }}
					class="flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary-600"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
					</svg>
					New Ticket
				</button>
			</div>

			{#if showNewTicket}
				<div class="mb-6 rounded-xl border border-primary-200 bg-primary-50/50 p-6 dark:border-primary-800/30 dark:bg-primary-900/10">
					{#if newTicketParentId}
						<h3 class="mb-4 text-sm font-semibold text-surface-900 dark:text-white">New Sub-task</h3>
					{/if}
					<form onsubmit={handleCreateTicket} class="flex flex-col gap-4">
						<input
							type="text"
							bind:value={newTicketTitle}
							placeholder="Ticket title"
							required
							class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						/>
						<textarea
							bind:value={newTicketDesc}
							placeholder="Description (optional)"
							rows="2"
							class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						></textarea>
						<select
							bind:value={newTicketPriority}
							class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						>
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
						<div class="flex gap-3">
							<button
								type="submit"
								disabled={creatingTicket}
								class="rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
							>
								{creatingTicket ? 'Creating...' : 'Create Ticket'}
							</button>
							<button
								type="button"
								onclick={() => { showNewTicket = false; newTicketParentId = null; }}
								class="rounded-lg px-4 py-2 text-sm font-medium text-surface-700 transition-colors hover:bg-surface-200 dark:text-surface-200 dark:hover:bg-surface-800"
							>
								Cancel
							</button>
						</div>
					</form>
				</div>
			{/if}

			{#snippet ticketCard(ticket: Ticket, depth: number)}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div 
					onclick={(e) => { e.stopPropagation(); if (editingTicketId !== ticket.id) toggleTicketExpanded(ticket.id); }}
					class="group cursor-pointer relative rounded-lg border bg-white p-4 shadow-sm transition-all dark:bg-surface-850
					{depth > 0 ? 'ml-6 mt-3 border-l-4 border-l-primary-400 border-y-surface-200/60 border-r-surface-200/60 dark:border-y-surface-800 dark:border-r-surface-800' : 'border-surface-200/60 hover:shadow-md dark:border-surface-800'}">
					
					{#if editingTicketId === ticket.id}
						<!-- Edit Form -->
						<div class="flex flex-col gap-3" onclick={(e) => e.stopPropagation()}>
							<input
								type="text"
								bind:value={editTicketTitle}
								placeholder="Ticket title"
								required
								class="rounded-lg border border-surface-200 bg-white px-3 py-1.5 text-sm font-medium text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
							/>
							<textarea
								bind:value={editTicketDesc}
								placeholder="Description (optional)"
								rows="2"
								class="rounded-lg border border-surface-200 bg-white px-3 py-1.5 text-xs text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
							></textarea>
							<div class="flex items-center justify-between">
								<select
									bind:value={editTicketPriority}
									class="rounded-lg border border-surface-200 bg-white px-2 py-1 text-xs text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
								>
									<option value="low">Low</option>
									<option value="medium">Medium</option>
									<option value="high">High</option>
									<option value="critical">Critical</option>
								</select>
								<div class="flex gap-2">
									<button
										type="button"
										onclick={cancelEditing}
										class="rounded px-2 py-1 text-xs font-medium text-surface-700 transition-colors hover:bg-surface-200 dark:text-surface-200 dark:hover:bg-surface-700"
									>
										Cancel
									</button>
									<button
										type="button"
										onclick={() => saveEdit(ticket)}
										disabled={isSavingEdit}
										class="rounded bg-primary-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
									>
										{isSavingEdit ? 'Saving...' : 'Save'}
									</button>
								</div>
							</div>
						</div>
					{:else}
						<!-- Normal Display -->
						<div class="mb-2 flex items-start justify-between">
							<h4 class="text-sm font-medium text-surface-900 dark:text-white">{ticket.title}</h4>
							<div class="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
								<button
									onclick={(e) => { e.stopPropagation(); startEditing(ticket); }}
									class="rounded p-1 text-surface-700/40 transition-colors hover:bg-surface-200 hover:text-surface-900 dark:text-surface-200/40 dark:hover:bg-surface-700 dark:hover:text-white"
									aria-label="Edit ticket"
								>
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
									</svg>
								</button>
								<button
									onclick={(e) => { e.stopPropagation(); handleDeleteTicket(ticket.id); }}
									class="rounded p-1 text-surface-700/40 transition-colors hover:bg-red-50 hover:text-red-500 dark:text-surface-200/40 dark:hover:bg-red-900/20 dark:hover:text-red-400"
									aria-label="Delete ticket"
								>
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
									</svg>
								</button>
							</div>
						</div>
						{#if ticket.description}
							<p class="mb-3 text-xs text-surface-700/50 dark:text-surface-200/30 whitespace-pre-wrap break-words {expandedTickets.includes(ticket.id) ? '' : 'line-clamp-2'}">{ticket.description}</p>
						{/if}
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-2">
								<span class="rounded-full px-2 py-0.5 text-[10px] font-semibold {PRIORITY_COLORS[ticket.priority] || PRIORITY_COLORS.medium}">
									{ticket.priority}
								</span>
								{#if depth < 3}
									<button 
										onclick={(e) => { e.stopPropagation(); newTicketParentId = ticket.id; showNewTicket = true; window.scrollTo({ top: 0, behavior: 'smooth' }); }}
										class="rounded px-1.5 py-0.5 text-[10px] font-medium text-surface-700/60 transition-colors hover:bg-surface-200 hover:text-surface-900 dark:text-surface-200/40 dark:hover:bg-surface-800 dark:hover:text-white"
										title="Add Sub-task"
									>
										+ Sub-task
									</button>
									<button
										type="button"
										onclick={(e) => { e.stopPropagation(); handleBreakdown(ticket.id); }}
										disabled={breakingDownId === ticket.id}
										class="flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium text-purple-600/80 transition-colors hover:bg-purple-50 hover:text-purple-700 dark:text-purple-400/80 dark:hover:bg-purple-900/20 dark:hover:text-purple-300 disabled:opacity-50"
										title="AI Breakdown"
									>
										<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.428-1.428L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.428-1.428l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.428 1.428l1.183.394-1.183.394a2.25 2.25 0 0 0-1.428 1.428Z" />
										</svg>
										{breakingDownId === ticket.id ? 'Thinking...' : 'AI Breakdown'}
									</button>
								{/if}
							</div>
							<!-- Move buttons -->
							<div class="flex gap-1">							{#if ticket.status !== 'todo'}
								<button
									onclick={(e) => { e.stopPropagation(); moveTicket(ticket, getLeftStatus(ticket.status)); }}
									class="rounded p-1 text-surface-700/30 transition-colors hover:bg-surface-200 hover:text-surface-700 dark:text-surface-200/20 dark:hover:bg-surface-800 dark:hover:text-surface-200"
									title="Move left"
								>
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
										<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
									</svg>
								</button>
							{/if}
							{#if ticket.status !== 'archived'}
								<button
									onclick={(e) => { e.stopPropagation(); moveTicket(ticket, getRightStatus(ticket.status)); }}
									class="rounded p-1 text-surface-700/30 transition-colors hover:bg-surface-200 hover:text-surface-700 dark:text-surface-200/20 dark:hover:bg-surface-800 dark:hover:text-surface-200"
									title="Move right"
								>
									<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
										<path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
									</svg>
								</button>
							{/if}
							</div>
						</div>
					{/if}

					<!-- Render sub-tasks -->
					{#each tickets.filter(t => t.parent_id === ticket.id) as child (child.id)}
						{@render ticketCard(child, depth + 1)}
					{/each}
				</div>
			{/snippet}

			<!-- Kanban board -->
			<div class="grid grid-cols-4 gap-6">
				{#each STATUSES as status}
					<div class="rounded-xl bg-surface-100/50 p-4 dark:bg-surface-900/50">
						<div class="mb-4 flex items-center justify-between">
							<div class="flex items-center gap-2">
								<span class="rounded-full px-2.5 py-1 text-xs font-semibold {STATUS_COLORS[status]}">
									{STATUS_LABELS[status]}
								</span>
								<span class="text-xs text-surface-700/40 dark:text-surface-200/30">{getColumnTickets(status).length}</span>
							</div>
						</div>
						<div class="space-y-3">
							{#each getColumnTickets(status) as ticket (ticket.id)}
								{@render ticketCard(ticket, 0)}
							{/each}
						</div>
					</div>
				{/each}
			</div>

	</div>
{:else}
	<div class="flex min-h-[50vh] items-center justify-center">
		<p class="text-sm text-surface-700/60 dark:text-surface-200/40 animate-pulse">Loading project...</p>
	</div>
{/if}
