<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { toast } from '$lib/stores/toast';
	import {
		getProject,
		listProjectMembers,
		removeProjectMember,
		updateProjectMemberRole,
		transferOwnership,
		listProjectInvites,
		createProjectInvite,
		deleteProjectInvite
	} from '$lib/api';
	import type { Project, ProjectMember, ProjectInvite } from '$lib/api';
	import { userStore } from '$lib/stores/user';
	import { projectMembersStore, wsStore } from '$lib/stores/websocket';

	let projectId = $derived($page.params.id);
	let user = $derived($userStore);
	let project = $state<Project | null>(null);
	let members = $derived($projectMembersStore[projectId] || []);
	let invites = $state<ProjectInvite[]>([]);
	let loading = $state(true);

	// Determine current user's role
	let currentUserRole = $derived.by(() => {
		if (!user) return null;
		const me = members.find((m) => m.user_id === user?.id);
		return me?.role || null;
	});

	let inviteEmail = $state('');
	let inviteRole = $state('member');
	let inviting = $state(false);

	onMount(async () => {
		try {
			const [proj, mems, invs] = await Promise.all([
				getProject(projectId),
				listProjectMembers(projectId),
				listProjectInvites(projectId).catch(() => []) // Admins/owners only
			]);
			project = proj;
			projectMembersStore.update((s) => ({ ...s, [projectId]: mems }));
			invites = invs;
		} catch (err) {
			console.error('Failed to load project members:', err);
			toast.error('Failed to load project members.');
		} finally {
			loading = false;
		}
	});

	async function handleInvite(e: SubmitEvent) {
		e.preventDefault();
		if (!inviteEmail.trim()) return;

		inviting = true;
		try {
			const invite = await createProjectInvite(projectId, inviteEmail, inviteRole);
			invites = [invite, ...invites];
			inviteEmail = '';
			toast.success('Invitation created!');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to invite');
		} finally {
			inviting = false;
		}
	}

	async function handleDeleteInvite(inviteId: string) {
		try {
			await deleteProjectInvite(projectId, inviteId);
			invites = invites.filter((i) => i.id !== inviteId);
			toast.success('Invite deleted');
		} catch (err) {
			toast.error('Failed to delete invite');
		}
	}

	async function handleRemoveMember(userId: string) {
		if (!confirm('Are you sure you want to remove this member?')) return;
		try {
			await removeProjectMember(projectId, userId);
			toast.success('Member removed');
			// WS will update local store
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to remove member');
		}
	}

	async function handleRoleChange(userId: string, newRole: string) {
		try {
			await updateProjectMemberRole(projectId, userId, newRole);
			toast.success('Role updated');
			// WS will update local store
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to update role');
		}
	}

	async function handleTransferOwnership(newOwnerId: string) {
		if (!confirm('Are you sure you want to transfer ownership? You will be demoted to Admin.')) return;
		try {
			await transferOwnership(projectId, newOwnerId);
			toast.success('Ownership transferred');
			// Refresh list entirely to reflect complex role shifts
			const mems = await listProjectMembers(projectId);
			projectMembersStore.update((s) => ({ ...s, [projectId]: mems }));
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to transfer ownership');
		}
	}
</script>

<svelte:head>
	<title>{project ? `Members - ${project.name}` : 'Members'} — NextUp</title>
</svelte:head>

<div class="max-w-5xl">
	{#if loading}
		<div class="mb-6 flex gap-4">
			<div class="h-9 w-32 animate-pulse rounded-lg bg-surface-200 dark:bg-surface-800"></div>
			<div class="h-9 w-56 animate-pulse rounded-lg bg-surface-200 dark:bg-surface-800"></div>
		</div>
	{:else if project}
		<!-- Header -->
		<div class="mb-6 flex items-center justify-between">
			<div>
				<a
					href="/projects/{projectId}"
					class="mb-2 inline-flex items-center gap-2 text-sm font-medium text-surface-500 transition-colors hover:text-surface-900 dark:text-surface-400 dark:hover:text-surface-100"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
					</svg>
					Back to Project
				</a>
				<h1 class="text-3xl font-bold tracking-tight text-surface-900 dark:text-white">
					Project Members
				</h1>
			</div>
		</div>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
			<div class="lg:col-span-2">
				<!-- Members List -->
				<div class="rounded-xl border border-surface-200/60 bg-white shadow-sm dark:border-surface-800 dark:bg-surface-900">
					<div class="border-b border-surface-200/60 px-6 py-4 dark:border-surface-800">
						<h2 class="font-semibold text-surface-900 dark:text-white">Current Members</h2>
					</div>
					<ul class="divide-y divide-surface-200/60 dark:divide-surface-800">
						{#each members as member}
							<li class="flex items-center justify-between px-6 py-4">
								<div class="flex items-center gap-3">
									<div class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-sm font-bold text-white">
										{member.first_name?.[0] || 'U'}
									</div>
									<div>
										<p class="font-medium text-surface-900 dark:text-white">
											{member.first_name} {member.last_name}
											{#if member.user_id === user?.id}
												<span class="ml-2 rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">You</span>
											{/if}
										</p>
										<p class="text-sm text-surface-500 dark:text-surface-400">{member.email}</p>
									</div>
								</div>
								
								<div class="flex items-center gap-3">
									<!-- Role display / select -->
									{#if currentUserRole === 'owner' && member.user_id !== user?.id}
										<select
											class="rounded-lg border border-surface-200 bg-surface-50 px-3 py-1.5 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
											value={member.role}
											onchange={(e) => handleRoleChange(member.user_id, e.currentTarget.value)}
										>
											<option value="admin">Admin</option>
											<option value="member">Member</option>
										</select>
									{:else}
										<span class="inline-flex items-center rounded-md px-2.5 py-1 text-sm font-medium capitalize 
											{member.role === 'owner' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' : ''}
											{member.role === 'admin' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : ''}
											{member.role === 'member' ? 'bg-surface-100 text-surface-700 dark:bg-surface-800 dark:text-surface-300' : ''}"
										>
											{member.role}
										</span>
									{/if}

									<!-- Actions -->
									{#if currentUserRole === 'owner' && member.role === 'admin'}
										<button
											onclick={() => handleTransferOwnership(member.user_id)}
											class="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
											title="Transfer Ownership"
										>
											Transfer
										</button>
									{/if}

									{#if (currentUserRole === 'owner' && member.user_id !== user?.id) || (currentUserRole === 'admin' && member.role === 'member')}
										<button
											onclick={() => handleRemoveMember(member.user_id)}
											class="rounded p-1.5 text-surface-400 hover:bg-red-50 hover:text-red-500 dark:text-surface-500 dark:hover:bg-red-900/20 dark:hover:text-red-400"
											title="Remove Member"
										>
											<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									{/if}
								</div>
							</li>
						{/each}
					</ul>
				</div>
			</div>

			<div class="lg:col-span-1 space-y-6">
				<!-- Invite Form -->
				{#if currentUserRole === 'owner' || currentUserRole === 'admin'}
					<div class="rounded-xl border border-surface-200/60 bg-white p-6 shadow-sm dark:border-surface-800 dark:bg-surface-900">
						<h2 class="mb-4 font-semibold text-surface-900 dark:text-white">Invite Member</h2>
						<form onsubmit={handleInvite} class="space-y-4">
							<div>
								<label class="mb-1 block text-sm font-medium text-surface-700 dark:text-surface-300">Email Address</label>
								<input
									type="email"
									bind:value={inviteEmail}
									required
									placeholder="colleague@example.com"
									class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3 py-2 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
								/>
							</div>
							<div>
								<label class="mb-1 block text-sm font-medium text-surface-700 dark:text-surface-300">Role</label>
								<select
									bind:value={inviteRole}
									class="w-full rounded-lg border border-surface-200 bg-surface-50 px-3 py-2 text-sm text-surface-900 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-surface-700 dark:bg-surface-800 dark:text-white"
								>
									<option value="member">Member</option>
									{#if currentUserRole === 'owner'}
										<option value="admin">Admin</option>
									{/if}
								</select>
							</div>
							<button
								type="submit"
								disabled={inviting}
								class="w-full rounded-lg bg-primary-500 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
							>
								{inviting ? 'Inviting...' : 'Send Invite'}
							</button>
						</form>
					</div>

					<!-- Pending Invites -->
					{#if invites.length > 0}
						<div class="rounded-xl border border-surface-200/60 bg-white p-6 shadow-sm dark:border-surface-800 dark:bg-surface-900">
							<h2 class="mb-4 font-semibold text-surface-900 dark:text-white">Pending Invites</h2>
							<ul class="space-y-3">
								{#each invites as invite}
									<li class="flex items-center justify-between rounded-lg border border-surface-100 bg-surface-50 p-3 dark:border-surface-700/50 dark:bg-surface-800/50">
										<div class="overflow-hidden">
											<p class="truncate text-sm font-medium text-surface-900 dark:text-white" title={invite.email}>
												{invite.email}
											</p>
											<p class="text-xs text-surface-500 dark:text-surface-400 capitalize">{invite.role}</p>
										</div>
										<button
											onclick={() => handleDeleteInvite(invite.id)}
											class="ml-2 p-1 text-surface-400 hover:text-red-500 dark:text-surface-500 dark:hover:text-red-400"
											title="Revoke Invite"
										>
											<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	{:else}
		<div class="flex min-h-[50vh] items-center justify-center">
			<p class="text-sm text-surface-700/60 dark:text-surface-200/40">Project not found.</p>
		</div>
	{/if}
</div>
