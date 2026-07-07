<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { userStore } from '$lib/stores/user';
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children } = $props();

	onMount(() => {
		const unsub = userStore.subscribe((user) => {
			if (user === null) {
				goto('/login');
			}
		});
		return unsub;
	});
</script>

<div class="flex min-h-screen bg-surface-50 dark:bg-surface-950">
	<Sidebar />
	<main class="ml-64 flex-1 p-8">
		{@render children()}
	</main>
</div>
