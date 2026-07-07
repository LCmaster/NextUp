<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { onMount } from 'svelte';
	import { userStore } from '$lib/stores/user';
	import { wsStore, applyWsEvent } from '$lib/stores/websocket';
	import Toast from '$lib/components/Toast.svelte';

	let { children } = $props();

	onMount(() => {
		// Attempt to restore the session from the JWT cookie.
		// We don't await here because onMount must return the cleanup function synchronously.
		// The restore() call is fire-and-forget; the store update is reactive.
		userStore.restore().then(() => {
			wsStore.connect();
		});

		const unsubscribe = wsStore.subscribe((event) => {
			if (event) applyWsEvent(event);
		});

		return () => {
			unsubscribe();
			wsStore.disconnect();
		};
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>NextUp — Task & Ticket Manager</title>
	<meta
		name="description"
		content="A modern ticketing and to-do list application for teams and individuals."
	/>
</svelte:head>

{@render children()}
<Toast />
