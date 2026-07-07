<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { acceptProjectInvite } from '$lib/api';
	import { userStore } from '$lib/stores/user';
	import { onMount } from 'svelte';

	let loading = $state(false);
	let error = $state('');
	let success = $state(false);
	
	const token = $page.params.token as string;

	onMount(() => {
		// If user is not logged in, they need to log in or register before accepting
		if (!$userStore) {
			// Ideally we would redirect back here after login, but for now just show a message or redirect
		}
	});

	async function handleAccept() {
		if (!$userStore) {
			goto('/login');
			return;
		}

		loading = true;
		error = '';
		try {
			await acceptProjectInvite(token);
			success = true;
			setTimeout(() => {
				goto('/dashboard');
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to accept invite';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-surface-100 via-primary-50 to-surface-100 dark:from-surface-950 dark:via-surface-900 dark:to-surface-950 p-4">
	<div class="w-full max-w-md text-center">
		<div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 text-2xl font-bold text-white shadow-xl shadow-primary-500/25">
			N
		</div>

		<div class="rounded-2xl border border-surface-200/60 bg-white/80 p-8 shadow-xl shadow-surface-900/5 backdrop-blur-sm dark:border-surface-800 dark:bg-surface-900/80">
			{#if success}
				<div class="mb-2 text-primary-500">
					<svg class="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h2 class="text-2xl font-bold text-surface-900 dark:text-white mb-2">Invite Accepted!</h2>
				<p class="text-surface-600 dark:text-surface-300">Redirecting to your dashboard...</p>
			{:else}
				<h1 class="text-2xl font-bold text-surface-900 dark:text-white mb-2">Project Invitation</h1>
				
				{#if !$userStore}
					<p class="mb-6 text-surface-600 dark:text-surface-300">
						You need to log in or create an account to accept this invitation.
					</p>
					<div class="space-y-3">
						<a href="/login" class="flex w-full items-center justify-center rounded-lg bg-gradient-to-r from-primary-500 to-primary-600 px-4 py-3 text-sm font-semibold text-white shadow-lg transition-all hover:from-primary-600 hover:to-primary-700">
							Log In
						</a>
						<a href="/register" class="flex w-full items-center justify-center rounded-lg border-2 border-primary-500 bg-transparent px-4 py-2.5 text-sm font-semibold text-primary-600 transition-all hover:bg-primary-50 dark:hover:bg-primary-900/20">
							Create Account
						</a>
					</div>
				{:else}
					<p class="mb-6 text-surface-600 dark:text-surface-300">
						You have been invited to join a project on NextUp.
					</p>
					
					{#if error}
						<div class="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
							{error}
						</div>
					{/if}

					<button
						onclick={handleAccept}
						disabled={loading}
						class="w-full rounded-lg bg-gradient-to-r from-primary-500 to-primary-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-primary-500/25 transition-all duration-200 hover:from-primary-600 hover:to-primary-700 hover:shadow-xl hover:shadow-primary-500/30 disabled:cursor-not-allowed disabled:opacity-50"
					>
						{loading ? 'Accepting...' : 'Accept Invitation'}
					</button>
				{/if}
			{/if}
		</div>
	</div>
</div>
