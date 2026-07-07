<script lang="ts">
	import TicketCard from '$lib/components/TicketCard.svelte';
	import { updateTicket, type Ticket } from '$lib/api';
	import { dndzone, type DndEvent } from 'svelte-dnd-action';

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

	// Local mutable copy of tickets for optimistic drag-and-drop
	let localTickets = $state<Ticket[]>([]);
	$effect(() => {
		// Sync when the upstream prop changes (e.g. from websockets)
		localTickets = [...tickets];
	});

	function getColumnTickets(status: string) {
		return localTickets.filter((t) => t.status === status && !t.parent_id);
	}

	function handleConsider(e: CustomEvent<DndEvent<Ticket>>, status: string) {
		const otherTickets = localTickets.filter((t) => t.status !== status || t.parent_id);
		// Force the status on the dragged items to the column they are currently hovering in
		const newColumnTickets = e.detail.items.map((t) => ({ ...t, status }));
		localTickets = [...otherTickets, ...newColumnTickets];
	}

	async function handleFinalize(e: CustomEvent<DndEvent<Ticket>>, status: string) {
		// Commit the local change
		handleConsider(e, status);

		const draggedId = e.detail.info.id;
		if (draggedId) {
			const ticket = localTickets.find((t) => t.id === draggedId);
			if (ticket) {
				try {
					await updateTicket(ticket.id, {
						title: ticket.title,
						priority: ticket.priority,
						status: ticket.status,
						description: ticket.description || undefined,
						assignee_id: ticket.assignee_id || undefined,
						parent_id: ticket.parent_id || undefined
					});
				} catch (err) {
					console.error('Failed to update ticket status after drop', err);
					// Revert on failure
					localTickets = [...tickets];
				}
			}
		}
	}
</script>

<div class="grid grid-cols-4 gap-6">
	{#each STATUSES as status}
		<div class="flex flex-col rounded-xl bg-surface-100/50 p-4 dark:bg-surface-900/50">
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
			<div
				class="flex-1 space-y-3 outline-none"
				use:dndzone={{
					items: getColumnTickets(status),
					flipDurationMs: 200,
					dropTargetStyle: {
						outline: '2px dashed var(--color-primary-500)',
						borderRadius: '0.5rem'
					}
				}}
				onconsider={(e) => handleConsider(e, status)}
				onfinalize={(e) => handleFinalize(e, status)}
			>
				{#each getColumnTickets(status) as ticket (ticket.id)}
					<div class="outline-none">
						<TicketCard {ticket} allTickets={localTickets} depth={0} {onrequestsubtask} />
					</div>
				{/each}
			</div>
		</div>
	{/each}
</div>
