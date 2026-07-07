<script lang="ts">
	// Heights for fake card placeholders — varied to look natural
	const COLUMNS = [
		{ cards: [72, 96, 72] },
		{ cards: [96, 72, 120, 72] },
		{ cards: [72, 96] },
		{ cards: [72] }
	];
</script>

<div class="grid grid-cols-4 gap-6" aria-busy="true" aria-label="Loading board…">
	{#each COLUMNS as col}
		<div class="rounded-xl bg-surface-100/50 p-4 dark:bg-surface-900/50">
			<!-- Column header shimmer -->
			<div class="mb-4 flex items-center gap-2">
				<div class="h-6 w-20 animate-pulse rounded-full bg-surface-200 dark:bg-surface-800"></div>
				<div class="h-4 w-5 animate-pulse rounded bg-surface-200 dark:bg-surface-800"></div>
			</div>

			<!-- Card shimmers -->
			<div class="space-y-3">
				{#each col.cards as height}
					<div
						class="rounded-lg border border-surface-200/60 bg-white p-4 shadow-sm dark:border-surface-800 dark:bg-surface-850"
						style="height: {height}px"
					>
						<div class="flex h-full flex-col gap-3">
							<!-- Title line -->
							<div class="h-4 w-3/4 animate-pulse rounded bg-surface-100 dark:bg-surface-800"></div>
							<!-- Body line -->
							{#if height > 72}
								<div
									class="h-3 w-full animate-pulse rounded bg-surface-100 dark:bg-surface-800"
								></div>
								<div
									class="h-3 w-2/3 animate-pulse rounded bg-surface-100 dark:bg-surface-800"
								></div>
							{/if}
							<!-- Footer chip -->
							<div
								class="mt-auto h-4 w-14 animate-pulse rounded-full bg-surface-100 dark:bg-surface-800"
							></div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/each}
</div>
