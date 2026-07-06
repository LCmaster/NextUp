<script lang="ts">
	import TicketCard from '$lib/components/TicketCard.svelte';
	import type { Ticket } from '$lib/api';

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

	interface Props {
		tickets: Ticket[];
		onrequestsubtask?: (parentId: string) => void;
	}

	let { tickets, onrequestsubtask }: Props = $props();

	/** Only top-level (non-child) tickets are rendered in each column; TicketCard renders children recursively. */
	function getColumnTickets(status: string) {
		return tickets.filter((t) => t.status === status && !t.parent_id);
	}
</script>

<div class="grid grid-cols-4 gap-6">
	{#each STATUSES as status}
		<div class="rounded-xl bg-surface-100/50 p-4 dark:bg-surface-900/50">
			<!-- Column header -->
			<div class="mb-4 flex items-center gap-2">
				<span class="rounded-full px-2.5 py-1 text-xs font-semibold {STATUS_COLORS[status]}">
					{STATUS_LABELS[status]}
				</span>
				<span class="text-xs text-surface-700/40 dark:text-surface-200/30">
					{getColumnTickets(status).length}
				</span>
			</div>

			<!-- Cards -->
			<div class="space-y-3">
				{#each getColumnTickets(status) as ticket (ticket.id)}
					<TicketCard
						{ticket}
						allTickets={tickets}
						depth={0}
						{onrequestsubtask}
					/>
				{/each}
			</div>
		</div>
	{/each}
</div>
