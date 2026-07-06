<script lang="ts">
	import { updateTicket, deleteTicket, breakdownTicket } from '$lib/api';
	import type { Ticket } from '$lib/api';

	const PRIORITY_COLORS: Record<string, string> = {
		low: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
		medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
		high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
		critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
	};

	interface Props {
		ticket: Ticket;
		allTickets: Ticket[];
		depth?: number;
		onrequestsubtask?: (parentId: string) => void;
		onbreakdown?: (ticketId: string) => void;
	}

	let {
		ticket,
		allTickets,
		depth = 0,
		onrequestsubtask,
		onbreakdown
	}: Props = $props();

	// ---- Local UI state ----
	let expanded = $state(false);
	let editing = $state(false);
	let breakingDown = $state(false);
	let saving = $state(false);

	// Edit field state
	let editTitle = $state('');
	let editDesc = $state('');
	let editPriority = $state('medium');

	function startEditing() {
		editTitle = ticket.title;
		editDesc = ticket.description || '';
		editPriority = ticket.priority;
		editing = true;
	}

	function cancelEditing() {
		editing = false;
	}

	async function saveEdit() {
		if (!editTitle.trim()) return;
		saving = true;
		try {
			await updateTicket(ticket.id, {
				title: editTitle.trim(),
				description: editDesc.trim() || undefined,
				status: ticket.status,
				priority: editPriority,
				assignee_id: ticket.assignee_id || undefined,
				parent_id: ticket.parent_id || undefined
			});
			editing = false;
		} catch (err) {
			console.error('Failed to update ticket:', err);
		} finally {
			saving = false;
		}
	}

	async function moveTicket(newStatus: string) {
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

	async function handleDelete() {
		try {
			await deleteTicket(ticket.id);
		} catch (err) {
			console.error('Failed to delete ticket:', err);
		}
	}

	async function handleBreakdown() {
		breakingDown = true;
		try {
			await breakdownTicket(ticket.id);
			onbreakdown?.(ticket.id);
		} catch (err) {
			console.error('Failed to breakdown ticket:', err);
		} finally {
			breakingDown = false;
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

	const children = $derived(allTickets.filter((t) => t.parent_id === ticket.id));
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	onclick={(e) => {
		e.stopPropagation();
		if (!editing) expanded = !expanded;
	}}
	class="group relative cursor-pointer rounded-lg border bg-white p-4 shadow-sm transition-all dark:bg-surface-850
		{depth > 0
		? 'ml-6 mt-3 border-l-4 border-l-primary-400 border-y-surface-200/60 border-r-surface-200/60 dark:border-y-surface-800 dark:border-r-surface-800'
		: 'border-surface-200/60 hover:shadow-md dark:border-surface-800'}"
>
	{#if editing}
		<!-- ── Edit Form ───────────────────────────────────────────────── -->
		<div class="flex flex-col gap-3" onclick={(e) => e.stopPropagation()}>
			<input
				type="text"
				bind:value={editTitle}
				placeholder="Ticket title"
				required
				class="rounded-lg border border-surface-200 bg-white px-3 py-1.5 text-sm font-medium text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
			/>
			<textarea
				bind:value={editDesc}
				placeholder="Description (optional)"
				rows="2"
				class="rounded-lg border border-surface-200 bg-white px-3 py-1.5 text-xs text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
			></textarea>
			<div class="flex items-center justify-between">
				<select
					bind:value={editPriority}
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
						onclick={saveEdit}
						disabled={saving}
						class="rounded bg-primary-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
					>
						{saving ? 'Saving...' : 'Save'}
					</button>
				</div>
			</div>
		</div>
	{:else}
		<!-- ── Normal Display ─────────────────────────────────────────── -->
		<div class="mb-2 flex items-start justify-between">
			<h4
				class="text-sm font-medium text-surface-900 dark:text-white {ticket.status === 'archived'
					? 'truncate'
					: ''}"
			>
				{ticket.title}
			</h4>

			<!-- Action buttons (visible on hover) -->
			<div class="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
				{#if ticket.status === 'archived'}
					<button
						onclick={(e) => { e.stopPropagation(); moveTicket(getLeftStatus(ticket.status)); }}
						class="rounded p-1 text-surface-700/40 transition-colors hover:bg-surface-200 hover:text-surface-900 dark:text-surface-200/40 dark:hover:bg-surface-700 dark:hover:text-white"
						aria-label="Revert to Done"
						title="Revert to Done"
					>
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
						</svg>
					</button>
				{:else}
					<button
						onclick={(e) => { e.stopPropagation(); startEditing(); }}
						class="rounded p-1 text-surface-700/40 transition-colors hover:bg-surface-200 hover:text-surface-900 dark:text-surface-200/40 dark:hover:bg-surface-700 dark:hover:text-white"
						aria-label="Edit ticket"
					>
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
						</svg>
					</button>
					<button
						onclick={(e) => { e.stopPropagation(); handleDelete(); }}
						class="rounded p-1 text-surface-700/40 transition-colors hover:bg-red-50 hover:text-red-500 dark:text-surface-200/40 dark:hover:bg-red-900/20 dark:hover:text-red-400"
						aria-label="Delete ticket"
					>
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
				{/if}
			</div>
		</div>

		<!-- Description (non-archived only) -->
		{#if ticket.description && ticket.status !== 'archived'}
			<p
				class="mb-3 whitespace-pre-wrap break-words text-xs text-surface-700/50 dark:text-surface-200/30 {expanded
					? ''
					: 'line-clamp-2'}"
			>
				{ticket.description}
			</p>
		{/if}

		<!-- Footer row (non-archived only) -->
		{#if ticket.status !== 'archived'}
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<span
						class="rounded-full px-2 py-0.5 text-[10px] font-semibold {PRIORITY_COLORS[ticket.priority] ||
							PRIORITY_COLORS.medium}"
					>
						{ticket.priority}
					</span>

					{#if depth < 3}
						<button
							onclick={(e) => { e.stopPropagation(); onrequestsubtask?.(ticket.id); }}
							class="rounded px-1.5 py-0.5 text-[10px] font-medium text-surface-700/60 transition-colors hover:bg-surface-200 hover:text-surface-900 dark:text-surface-200/40 dark:hover:bg-surface-800 dark:hover:text-white"
							title="Add Sub-task"
						>
							+ Sub-task
						</button>

						<button
							type="button"
							onclick={(e) => { e.stopPropagation(); handleBreakdown(); }}
							disabled={breakingDown}
							class="flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium text-purple-600/80 transition-colors hover:bg-purple-50 hover:text-purple-700 dark:text-purple-400/80 dark:hover:bg-purple-900/20 dark:hover:text-purple-300 disabled:opacity-50"
							title="AI Breakdown"
						>
							<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.428-1.428L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.428-1.428l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.428 1.428l1.183.394-1.183.394a2.25 2.25 0 0 0-1.428 1.428Z" />
							</svg>
							{breakingDown ? 'Thinking...' : 'AI Breakdown'}
						</button>
					{/if}
				</div>

				<!-- Move buttons -->
				<div class="flex gap-1">
					{#if ticket.status !== 'todo'}
						<button
							onclick={(e) => { e.stopPropagation(); moveTicket(getLeftStatus(ticket.status)); }}
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
							onclick={(e) => { e.stopPropagation(); moveTicket(getRightStatus(ticket.status)); }}
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
	{/if}

	<!-- Sub-tasks (recursive) -->
	{#each children as child (child.id)}
		<svelte:self
			ticket={child}
			{allTickets}
			depth={depth + 1}
			{onrequestsubtask}
			{onbreakdown}
		/>
	{/each}
</div>
