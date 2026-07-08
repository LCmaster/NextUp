<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { userStore } from '$lib/stores/user';

	let loading = $state(true);

	onMount(() => {
		// If API is not reachable after a timeout, show fallback
		const timeout = setTimeout(() => {
			if ($userStore === undefined) {
				loading = false;
			}
		}, 3000);

		return () => {
			clearTimeout(timeout);
		};
	});
</script>

{#if loading}
	<div class="flex min-h-screen items-center justify-center bg-surface-50 dark:bg-surface-950">
		<div class="text-center">
			<div
				class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 text-2xl font-bold text-white shadow-xl shadow-primary-500/25 animate-pulse"
			>
				N
			</div>
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">Loading NextUp...</p>
		</div>
	</div>
{:else}
	<div class="flex min-h-screen items-center justify-center bg-surface-50 dark:bg-surface-950">
		<div class="text-center">
			<div
				class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 text-2xl font-bold text-white shadow-xl shadow-primary-500/25"
			>
				N
			</div>
			<h1 class="mb-2 text-2xl font-bold text-surface-900 dark:text-white">Unable to connect</h1>
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">
				Make sure the backend server is running.
			</p>
		</div>
	</div>
{/if}
