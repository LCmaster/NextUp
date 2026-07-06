<script lang="ts">
	import { createTicket } from '$lib/api';
	import { ticketsStore } from '$lib/stores/websocket';
	import type { Ticket } from '$lib/api';

	interface Props {
		projectId: string;
		parentId?: string | null;
		oncancel: () => void;
		oncreated?: (ticket: Ticket) => void;
	}

	let { projectId, parentId = null, oncancel, oncreated }: Props = $props();

	let title = $state('');
	let description = $state('');
	let priority = $state('medium');
	let creating = $state(false);
	let error = $state('');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!title.trim()) return;

		creating = true;
		error = '';
		try {
			const ticket = await createTicket({
				project_id: projectId,
				title: title.trim(),
				description: description.trim() || undefined,
				priority,
				parent_id: parentId || undefined
			});
			title = '';
			description = '';
			priority = 'medium';
			oncreated?.(ticket);
			oncancel();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create ticket';
		} finally {
			creating = false;
		}
	}
</script>

<div class="mb-6 rounded-xl border border-primary-200 bg-primary-50/50 p-6 dark:border-primary-800/30 dark:bg-primary-900/10">
	{#if parentId}
		<h3 class="mb-4 text-sm font-semibold text-surface-900 dark:text-white">New Sub-task</h3>
	{/if}

	{#if error}
		<p class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">
			{error}
		</p>
	{/if}

	<form onsubmit={handleSubmit} class="flex flex-col gap-4">
		<input
			type="text"
			bind:value={title}
			placeholder="Ticket title"
			required
			class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
		/>
		<textarea
			bind:value={description}
			placeholder="Description (optional)"
			rows="2"
			class="rounded-lg border border-surface-200 bg-white px-3.5 py-2.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
		></textarea>
		<select
			bind:value={priority}
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
				disabled={creating}
				class="rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
			>
				{creating ? 'Creating...' : 'Create Ticket'}
			</button>
			<button
				type="button"
				onclick={oncancel}
				class="rounded-lg px-4 py-2 text-sm font-medium text-surface-700 transition-colors hover:bg-surface-200 dark:text-surface-200 dark:hover:bg-surface-800"
			>
				Cancel
			</button>
		</div>
	</form>
</div>
