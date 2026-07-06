<script lang="ts">
	import { goto } from '$app/navigation';
	import { setupAccount } from '$lib/api';
	import { userStore } from '$lib/stores/user';

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let githubLink = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}
		if (password.length < 6) {
			error = 'Password must be at least 6 characters';
			return;
		}

		loading = true;
		try {
			const user = await setupAccount({
				first_name: firstName,
				last_name: lastName,
				email,
				password,
				github_link: githubLink || undefined
			});
			userStore.login(user);
			goto('/dashboard');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Setup failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-surface-100 via-primary-50 to-surface-100 dark:from-surface-950 dark:via-surface-900 dark:to-surface-950 p-4">
	<div class="w-full max-w-md">
		<!-- Logo -->
		<div class="mb-8 text-center">
			<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 text-2xl font-bold text-white shadow-xl shadow-primary-500/25">
				N
			</div>
			<h1 class="text-3xl font-bold tracking-tight text-surface-900 dark:text-white">Welcome to NextUp</h1>
			<p class="mt-2 text-surface-700/70 dark:text-surface-200/50">Set up your account to get started</p>
		</div>

		<!-- Form card -->
		<div class="rounded-2xl border border-surface-200/60 bg-white/80 p-8 shadow-xl shadow-surface-900/5 backdrop-blur-sm dark:border-surface-800 dark:bg-surface-900/80">
			{#if error}
				<div class="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
					{error}
				</div>
			{/if}

			<form onsubmit={handleSubmit} class="space-y-5">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label for="firstName" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
							First Name
						</label>
						<input
							id="firstName"
							type="text"
							bind:value={firstName}
							required
							class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
							placeholder="John"
						/>
					</div>
					<div>
						<label for="lastName" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
							Last Name
						</label>
						<input
							id="lastName"
							type="text"
							bind:value={lastName}
							required
							class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
							placeholder="Doe"
						/>
					</div>
				</div>

				<div>
					<label for="email" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
						Email
					</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						placeholder="john@example.com"
					/>
				</div>

				<div>
					<label for="password" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
						Password
					</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						minlength="6"
						class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						placeholder="••••••••"
					/>
				</div>

				<div>
					<label for="confirmPassword" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
						Confirm Password
					</label>
					<input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						required
						minlength="6"
						class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						placeholder="••••••••"
					/>
				</div>

				<div>
					<label for="githubLink" class="mb-1.5 block text-sm font-medium text-surface-700 dark:text-surface-200">
						GitHub Profile <span class="text-surface-700/40 dark:text-surface-200/30">(optional)</span>
					</label>
					<input
						id="githubLink"
						type="url"
						bind:value={githubLink}
						class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3.5 py-2.5 text-sm text-surface-900 transition-colors focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
						placeholder="https://github.com/username"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-lg bg-gradient-to-r from-primary-500 to-primary-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-primary-500/25 transition-all duration-200 hover:from-primary-600 hover:to-primary-700 hover:shadow-xl hover:shadow-primary-500/30 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{loading ? 'Creating Account...' : 'Create Account'}
				</button>
			</form>
		</div>
	</div>
</div>
