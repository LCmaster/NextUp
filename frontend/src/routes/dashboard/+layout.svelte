<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children } = $props();
</script>

<div class="flex min-h-screen bg-surface-50 dark:bg-surface-950">
	<Sidebar />
	<main class="ml-64 flex-1 p-8">
		{@render children()}
	</main>
</div>
