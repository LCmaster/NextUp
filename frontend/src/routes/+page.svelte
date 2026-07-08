<script lang="ts">
	import { onMount } from 'svelte';
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
			<img src="/logo-light.png" alt="NextUp Logo" class="mx-auto mb-4 h-16 object-contain animate-pulse dark:hidden" />
			<img src="/logo-dark.png" alt="NextUp Logo" class="mx-auto mb-4 hidden h-16 object-contain animate-pulse dark:block" />
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">Loading NextUp...</p>
		</div>
	</div>
{:else}
	<div class="flex min-h-screen items-center justify-center bg-surface-50 dark:bg-surface-950">
		<div class="text-center">
			<img src="/logo-light.png" alt="NextUp Logo" class="mx-auto mb-4 h-16 object-contain dark:hidden" />
			<img src="/logo-dark.png" alt="NextUp Logo" class="mx-auto mb-4 hidden h-16 object-contain dark:block" />
			<h1 class="mb-2 text-2xl font-bold text-surface-900 dark:text-white">Unable to connect</h1>
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">
				Make sure the backend server is running.
			</p>
		</div>
	</div>
{/if}
