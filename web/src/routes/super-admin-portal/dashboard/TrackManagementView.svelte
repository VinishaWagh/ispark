<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { API_BASE_URL } from '$lib/config';

	// ── Types ────────────────────────────────────────────────────────────────
	type TrackStatus = 'Active' | 'Inactive';

	interface Track {
		id: number;
		name: string;
		description: string;
		totalActivities: number;
		status: TrackStatus;
	}

	// ── Track Registry (loaded from the API) ─────────────────────────────────
	let tracks = $state<Track[]>([]);

	const tracksBase = `${API_BASE_URL}/api/admin/platform/tracks`;

	function authHeaders(): Record<string, string> {
		return { Authorization: `Bearer ${localStorage.getItem('superadmin_token')}` };
	}

	async function unauthorized(res: Response): Promise<boolean> {
		if (res.status === 401) {
			localStorage.removeItem('superadmin_token');
			await goto('/super-admin-portal');
			return true;
		}
		return false;
	}

	async function loadTracks() {
		try {
			const res = await fetch(tracksBase, { headers: authHeaders() });
			if (await unauthorized(res)) return;
			if (!res.ok) {
				triggerToast('Could not load tracks. Please try again.');
				return;
			}
			const data = await res.json();
			const loaded = data.tracks || [];
			tracks = loaded.map(
				(t: {
					id: number;
					name: string;
					description: string;
					status: TrackStatus;
					totalActivities?: number;
					total_activities?: number;
				}) => ({
					id: t.id,
					name: t.name,
					description: t.description,
					totalActivities: t.totalActivities ?? t.total_activities ?? 0,
					status: t.status
				})
			);
		} catch {
			triggerToast('Could not load tracks. Please try again.');
		}
	}

	onMount(loadTracks);

	let trackFilter = $state('All');
	let trackSearch = $state('');

	let filteredTracks = $derived(
		tracks.filter((t) => {
			const matchesFilter = trackFilter === 'All' || t.status === trackFilter;
			const matchesSearch =
				t.name.toLowerCase().includes(trackSearch.toLowerCase()) ||
				t.description.toLowerCase().includes(trackSearch.toLowerCase());
			return matchesFilter && matchesSearch;
		})
	);

	// ── Stat card derivations ────────────────────────────────────────────────
	let totalTracksCount = $derived(tracks.length);
	let activeTracksCount = $derived(tracks.filter((t) => t.status === 'Active').length);
	let personalityDevActivities = $derived(
		tracks.find((t) => t.name === 'Personality Development')?.totalActivities ?? 0
	);
	let skillBuildingActivities = $derived(
		tracks.find((t) => t.name === 'Skill Building')?.totalActivities ?? 0
	);

	// ── Add Track modal ──────────────────────────────────────────────────────
	let isAddTrackModalOpen = $state(false);
	let newTrackName = $state('');
	let newTrackDescription = $state('');

	async function handleAddTrack(e: Event) {
		e.preventDefault();
		const name = newTrackName.trim();
		if (!name) return;

		try {
			const res = await fetch(tracksBase, {
				method: 'POST',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({ name, description: newTrackDescription.trim() })
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to create track');
				return;
			}

			triggerToast(`Track "${name}" created successfully!`);
			newTrackName = '';
			newTrackDescription = '';
			isAddTrackModalOpen = false;
			await loadTracks();
		} catch {
			triggerToast('Failed to create track');
		}
	}

	// ── Edit Track modal ─────────────────────────────────────────────────────
	let isEditTrackModalOpen = $state(false);
	let editTrackId = $state(-1);
	let editTrackName = $state('');
	let editTrackDescription = $state('');
	let editTrackStatus = $state<TrackStatus>('Active');

	function openEditTrack(track: Track) {
		editTrackId = track.id;
		editTrackName = track.name;
		editTrackDescription = track.description;
		editTrackStatus = track.status;
		isEditTrackModalOpen = true;
	}

	async function handleSaveTrack(e: Event) {
		e.preventDefault();
		const name = editTrackName.trim();
		if (editTrackId < 0 || !name) return;

		try {
			const res = await fetch(`${tracksBase}/${editTrackId}`, {
				method: 'PUT',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name,
					description: editTrackDescription.trim(),
					status: editTrackStatus
				})
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to update track');
				return;
			}

			triggerToast(`Track "${name}" updated successfully!`);
			isEditTrackModalOpen = false;
			editTrackId = -1;
			await loadTracks();
		} catch {
			triggerToast('Failed to update track');
		}
	}

	// ── View Track modal ─────────────────────────────────────────────────────
	let isViewTrackModalOpen = $state(false);
	let viewTrack = $state<Track | null>(null);

	function openViewTrack(track: Track) {
		viewTrack = track;
		isViewTrackModalOpen = true;
	}

	// ── Toast notifications ──────────────────────────────────────────────────
	interface Toast {
		id: number;
		message: string;
	}
	let toasts = $state<Toast[]>([]);
	let toastCounter = 0;

	function triggerToast(message: string) {
		const id = toastCounter++;
		toasts = [...toasts, { id, message }];
		setTimeout(() => {
			toasts = toasts.filter((t) => t.id !== id);
		}, 3000);
	}

	async function handleDeleteTrack(track: Track) {
		if (!confirm(`Are you sure you want to delete the "${track.name}" track?`)) return;

		try {
			const res = await fetch(`${tracksBase}/${track.id}`, {
				method: 'DELETE',
				headers: authHeaders()
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to delete track');
				return;
			}

			triggerToast(`Track "${track.name}" removed successfully.`);
			await loadTracks();
		} catch {
			triggerToast('Failed to delete track');
		}
	}

	function trackStatusClass(status: TrackStatus): string {
		return status === 'Active' ? 'text-emerald-600' : 'text-slate-400';
	}

	function trackStatusDot(status: TrackStatus): string {
		return status === 'Active' ? 'bg-emerald-600' : 'bg-slate-400';
	}
</script>

<!-- ==================== TRACK MANAGEMENT ==================== -->
<div class="space-y-6">
	<!-- Overview Stat Cards -->
	<section
		class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
		aria-label="Track management overview"
	>
		<!-- Total Tracks -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{totalTracksCount}</span>
				<div class="p-2.5 rounded-lg bg-slate-100 text-slate-600 border border-slate-200">
					<!-- Layers Icon for Total Tracks -->
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M12 3 2.25 8.25 12 13.5l9.75-5.25L12 3Z"
						/>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="m2.25 12.75 9.75 5.25 9.75-5.25"
						/>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="m2.25 17.25 9.75 5.25 9.75-5.25"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Total tracks</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Total platform tracks
				</p>
			</div>
		</div>

		<!-- Active Tracks -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{activeTracksCount}</span>
				<div class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100">
					<!-- Trending Up / Pulse Icon -->
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M2.25 18 9 11.25l4.306 4.306a11.95 11.95 0 015.814-5.518l2.74-1.22m0 0-5.94-2.281m5.94 2.28-2.28 5.941"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Active tracks</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Currently active tracks
				</p>
			</div>
		</div>

		<!-- Personality Development Activities -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{personalityDevActivities}</span>
				<div class="p-2.5 rounded-lg bg-purple-50 text-purple-600 border border-purple-100">
					<!-- User Development Icon -->
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
					Personality development
				</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					{personalityDevActivities} registered activities
				</p>
			</div>
		</div>

		<!-- Skill Building Activities -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{skillBuildingActivities}</span>
				<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100">
					<!-- Lightbulb Icon -->
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M12 18v-1.5m0 0a6 6 0 10-6-6c0 2.22 1.21 4.156 3 5.19V16.5a1.5 1.5 0 001.5 1.5h3a1.5 1.5 0 001.5-1.5v-1.01a6.002 6.002 0 003-5.19c0-3.314-2.686-6-6-6zM9.75 21h4.5"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Skill building</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					{skillBuildingActivities} registered activities
				</p>
			</div>
		</div>
	</section>

	<!-- Track Management Overview -->
	<section class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
		<!-- Header -->
		<div
			class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-slate-50/20 select-none"
		>
			<div>
				<h3 class="text-base font-bold font-serif text-slate-900">Track Management Overview</h3>
				<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
					{filteredTracks.length} of {tracks.length} tracks
				</p>
			</div>
			<button
				type="button"
				onclick={() => (isAddTrackModalOpen = true)}
				class="w-full sm:w-auto inline-flex items-center justify-center gap-1.5 px-4 py-2.5 sm:py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none shrink-0"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="2.5"
					stroke="currentColor"
					class="w-3.5 h-3.5"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Add Track
			</button>
		</div>

		<!-- Filters & Search -->
		<div
			class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white select-none"
		>
			<div class="flex flex-wrap gap-1.5">
				{#each ['All', 'Active', 'Inactive'] as filterOption}
					<button
						type="button"
						onclick={() => (trackFilter = filterOption)}
						class="px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all
							{trackFilter === filterOption
							? 'bg-[#881B1B] text-white shadow-xs'
							: 'bg-slate-50 text-slate-500 hover:bg-slate-100'}"
					>
						{filterOption}
					</button>
				{/each}
			</div>

			<div class="relative w-full sm:w-64">
				<input
					type="text"
					bind:value={trackSearch}
					placeholder="Search track name..."
					class="pl-4 pr-9 py-2 bg-slate-50 rounded-lg border border-slate-200 text-xs text-slate-800 focus:outline-none focus:border-slate-350 focus:bg-white w-full transition-all"
				/>
				<span class="absolute right-3 top-2.5 text-slate-400">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-4 h-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
						/>
					</svg>
				</span>
			</div>
		</div>

		<!-- Table -->
		<div class="overflow-x-auto no-scrollbar">
			<table class="w-full text-left border-collapse min-w-[600px]">
				<thead>
					<tr
						class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
					>
						<th class="py-3.5 px-5">Track Name</th>
						<th class="py-3.5 px-5">Description</th>
						<th class="py-3.5 px-5">Total Activities</th>
						<th class="py-3.5 px-5">Status</th>
						<th class="py-3.5 px-5 text-center">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100 text-xs font-sans">
					{#if filteredTracks.length === 0}
						<tr>
							<td colspan="5" class="py-8 text-center text-slate-400 font-semibold select-none">
								No tracks found matching search filters.
							</td>
						</tr>
					{:else}
						{#each filteredTracks as track (track.id)}
							<tr class="hover:bg-slate-50/30 transition-colors">
								<td class="py-4 px-5 font-bold text-slate-800 align-top whitespace-nowrap">
									{track.name}
								</td>
								<td class="py-4 px-5 text-slate-500 font-semibold align-top max-w-sm">
									{track.description}
								</td>
								<td class="py-4 px-5 font-bold text-slate-800 align-top">
									{track.totalActivities}
								</td>
								<td class="py-4 px-5 align-top">
									<span
										class="inline-flex items-center gap-1.5 font-bold {trackStatusClass(
											track.status
										)}"
									>
										<span class="w-1.5 h-1.5 rounded-full shrink-0 {trackStatusDot(track.status)}"
										></span>
										{track.status}
									</span>
								</td>
								<td class="py-4 px-5 align-top">
									<div class="flex items-center justify-center gap-3 text-slate-400">
										<button
											type="button"
											onclick={() => openViewTrack(track)}
											aria-label="View track"
											class="text-blue-500 hover:text-blue-700 transition-colors p-0.5"
										>
											<svg
												xmlns="http://www.w3.org/2000/svg"
												fill="none"
												viewBox="0 0 24 24"
												stroke-width="2"
												stroke="currentColor"
												class="w-4 h-4"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z"
												/>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
												/>
											</svg>
										</button>
										<button
											type="button"
											onclick={() => openEditTrack(track)}
											aria-label="Edit track"
											class="text-amber-500 hover:text-amber-700 transition-colors p-0.5"
										>
											<svg
												xmlns="http://www.w3.org/2000/svg"
												fill="none"
												viewBox="0 0 24 24"
												stroke-width="2"
												stroke="currentColor"
												class="w-4 h-4"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125"
												/>
											</svg>
										</button>
										<button
											type="button"
											onclick={() => handleDeleteTrack(track)}
											aria-label="Delete track"
											class="text-rose-500 hover:text-rose-700 transition-colors p-0.5"
										>
											<svg
												xmlns="http://www.w3.org/2000/svg"
												fill="none"
												viewBox="0 0 24 24"
												stroke-width="2"
												stroke="currentColor"
												class="w-4 h-4"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
												/>
											</svg>
										</button>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		<!-- Footer -->
		<div
			class="p-4 border-t border-slate-100 bg-slate-50/30 text-slate-500 font-semibold text-[11px] select-none"
		>
			<span>Showing {filteredTracks.length} of {tracks.length} tracks</span>
		</div>
	</section>
</div>

<!-- ==================== TOAST NOTIFICATIONS ==================== -->
<div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2 max-w-sm">
	{#each toasts as toast (toast.id)}
		<div
			transition:slide={{ duration: 150 }}
			class="p-4 bg-slate-800 border border-slate-700 text-white rounded-xl shadow-2xl flex items-center gap-2 text-xs font-semibold font-sans"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2"
				stroke="currentColor"
				class="w-4 h-4 text-emerald-400"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M9 12.75 11.25 15 15 9.75M21 12c0 1.268-.63 2.39-1.593 3.068a3.745 3.745 0 01-1.043 3.296 3.745 3.745 0 01-3.296 1.043A3.745 3.745 0 0112 21c-1.268 0-2.39-.63-3.068-1.593a3.746 3.746 0 01-3.296-1.043 3.745 3.745 0 01-1.043-3.296A3.745 3.745 0 013 12c0-1.268.63-2.39 1.593-3.068a3.745 3.745 0 011.043-3.296 3.746 3.746 0 013.296-1.043A3.746 3.746 0 0112 3c1.268 0 2.39.63 3.068 1.593a3.746 3.746 0 013.296 1.043 3.746 3.746 0 011.043 3.296A3.745 3.745 0 0121 12Z"
				/>
			</svg>
			<span>{toast.message}</span>
		</div>
	{/each}
</div>

<!-- ==================== MODALS ==================== -->

<!-- Add Track Modal -->
{#if isAddTrackModalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isAddTrackModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<form
			onsubmit={handleAddTrack}
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">Add New Track</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Register a platform track
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isAddTrackModalOpen = false)}
					aria-label="Close modal"
					class="p-1 rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-6 space-y-4">
				<div class="flex flex-col gap-1.5">
					<label
						for="new-track-name"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">TRACK NAME *</label
					>
					<input
						id="new-track-name"
						type="text"
						bind:value={newTrackName}
						placeholder="e.g. Social Entrepreneurship & Innovation"
						required
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<label
						for="new-track-description"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">DESCRIPTION</label
					>
					<textarea
						id="new-track-description"
						bind:value={newTrackDescription}
						rows="3"
						placeholder="Briefly describe the focus of this track"
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 resize-none"
					></textarea>
				</div>
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
			>
				<button
					type="button"
					onclick={() => (isAddTrackModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Cancel
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Add Track
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- Edit Track Modal -->
{#if isEditTrackModalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isEditTrackModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<form
			onsubmit={handleSaveTrack}
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">Edit Track</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Update track details
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isEditTrackModalOpen = false)}
					aria-label="Close modal"
					class="p-1 rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-6 space-y-4">
				<div class="flex flex-col gap-1.5">
					<label
						for="edit-track-name"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">TRACK NAME *</label
					>
					<input
						id="edit-track-name"
						type="text"
						bind:value={editTrackName}
						required
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<label
						for="edit-track-description"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">DESCRIPTION</label
					>
					<textarea
						id="edit-track-description"
						bind:value={editTrackDescription}
						rows="3"
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 resize-none"
					></textarea>
				</div>

				<div class="flex flex-col gap-1.5">
					<label
						for="edit-track-status"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">STATUS</label
					>
					<select
						id="edit-track-status"
						bind:value={editTrackStatus}
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 bg-white focus:outline-none focus:border-slate-355"
					>
						<option value="Active">Active</option>
						<option value="Inactive">Inactive</option>
					</select>
				</div>
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
			>
				<button
					type="button"
					onclick={() => (isEditTrackModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Cancel
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Save Changes
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- View Track Modal -->
{#if isViewTrackModalOpen && viewTrack}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isViewTrackModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<div
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans max-h-[90vh]"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">View Track</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Track Details ({viewTrack.id})
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isViewTrackModalOpen = false)}
					aria-label="Close modal"
					class="p-1 rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						class="w-5 h-5"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-6 space-y-4 overflow-y-auto flex-grow">
				<!-- Track Name -->
				<div class="flex flex-col gap-1.5">
					<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
						TRACK NAME
					</span>
					<div
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
					>
						{viewTrack.name}
					</div>
				</div>

				<!-- Description -->
				<div class="flex flex-col gap-1.5">
					<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
						DESCRIPTION
					</span>
					<div
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50 min-h-[60px] leading-relaxed"
					>
						{viewTrack.description || 'No description provided.'}
					</div>
				</div>

				<!-- Total Activities & Status -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							TOTAL ACTIVITIES
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewTrack.totalActivities}
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							STATUS
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-bold bg-slate-50/50 flex items-center gap-1.5 {trackStatusClass(
								viewTrack.status
							)}"
						>
							<span class="w-1.5 h-1.5 rounded-full shrink-0 {trackStatusDot(viewTrack.status)}"
							></span>
							{viewTrack.status}
						</div>
					</div>
				</div>
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30 shrink-0"
			>
				<button
					type="button"
					onclick={() => (isViewTrackModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Close
				</button>
				<button
					type="button"
					onclick={() => {
						isViewTrackModalOpen = false;
						if (viewTrack) openEditTrack(viewTrack);
					}}
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Edit Track
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.no-scrollbar {
		scrollbar-width: none; /* Firefox */
		-ms-overflow-style: none; /* IE / Edge legacy */
	}
	.no-scrollbar::-webkit-scrollbar {
		display: none; /* Chrome, Safari, Opera */
		width: 0;
		height: 0;
	}
</style>
