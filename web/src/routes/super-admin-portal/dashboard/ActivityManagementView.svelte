<script lang="ts">
	import { fade, slide } from 'svelte/transition';

	// ── Types ───────────────────────────────────────────────────────────────────
	type Track = string;
	type ActivityType = 'Workshop' | 'Seminar' | 'Webinar' | 'Course';
	type ActivityStatus = 'Active' | 'Inactive';

	interface TrackOption {
		id: number;
		name: string;
	}

	interface Activity {
		id: string;
		name: string;
		track: Track;
		type: ActivityType;
		credits: number;
		status: ActivityStatus;
		category?: string;
		description?: string;
		mode?: string;
		reg_deadline?: string;
		activity_date?: string;
		venue?: string;
		coordinator?: string;
	}

	import { API_BASE_URL } from '$lib/config';

	let activities = $state<Activity[]>([]);
	let availableTracks = $state<TrackOption[]>([]);

	// ── Credit Rules ────────────────────────────────────────────────────────────
	let creditRules = $state<Record<ActivityType, number>>({
		Workshop: 2,
		Seminar: 2,
		Webinar: 2,
		Course: 4
	});

	async function loadActivities() {
		try {
			const token = localStorage.getItem('superadmin_token') || '';
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/activities`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});
			if (res.ok) {
				const data = await res.json();
				activities = data.activities || [];
			} else {
				triggerToast('Failed to load activities');
			}
		} catch (err) {
			console.error(err);
			triggerToast('Network error while loading activities');
		}
	}

	async function loadTracks() {
		try {
			const token = localStorage.getItem('superadmin_token') || '';
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/tracks`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});
			if (res.ok) {
				const data = await res.json();
				availableTracks = data.tracks || [];
			}
		} catch (err) {
			console.error(err);
		}
	}

	$effect(() => {
		loadActivities();
		loadTracks();
	});

	// ── Filters / Search / Pagination ───────────────────────────────────────────
	type FilterTab = string;
	let activeFilter = $state<FilterTab>('All');
	let filterTabs = $derived(['All', ...availableTracks.map((t) => t.name), 'Active', 'Inactive']);
	let searchQuery = $state('');
	let currentPage = $state(1);
	const pageSize = 10;

	let filteredActivities = $derived(
		activities.filter((a) => {
			const matchesFilter =
				activeFilter === 'All' || a.track === activeFilter || a.status === activeFilter;
			const matchesSearch =
				a.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				a.type.toLowerCase().includes(searchQuery.toLowerCase()) ||
				a.id.toLowerCase().includes(searchQuery.toLowerCase());
			return matchesFilter && matchesSearch;
		})
	);

	let totalPages = $derived(Math.max(1, Math.ceil(filteredActivities.length / pageSize)));

	let paginatedActivities = $derived(
		filteredActivities.slice((currentPage - 1) * pageSize, currentPage * pageSize)
	);

	function setFilter(tab: FilterTab) {
		activeFilter = tab;
		currentPage = 1;
	}

	function setSearch(value: string) {
		searchQuery = value;
		currentPage = 1;
	}

	function goToPage(page: number) {
		if (page < 1 || page > totalPages) return;
		currentPage = page;
	}

	// ── Stats (Derived) ──────────────────────────────────────────────────────────
	let totalActivitiesCount = $derived(activities.length);
	let activeActivitiesCount = $derived(activities.filter((a) => a.status === 'Active').length);
	let personalityDevCount = $derived(
		activities.filter((a) => a.track === 'Personality Development').length
	);
	let skillBuildingCount = $derived(activities.filter((a) => a.track === 'Skill Building').length);

	// ── Toast Notifications ──────────────────────────────────────────────────────
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

	// ── Add / Edit Activity Modal ────────────────────────────────────────────────
	let {
		isActivityModalOpen = $bindable(false)
	}: {
		isActivityModalOpen?: boolean;
	} = $props();
	let modalMode = $state<'add' | 'edit'>('add');
	let editingId = $state<string | null>(null);

	let formName = $state('');
	let formTrack = $state<Track>('');
	let userSelectedTrack = $state(false);
	let formType = $state<ActivityType>('Workshop');
	let formCredits = $state(2);
	let formStatus = $state<ActivityStatus>('Active');

	let formCategory = $state('TECHNICAL');
	let formDescription = $state('');
	let formMode = $state('Offline');
	let formRegDeadline = $state('');
	let formActivityDate = $state('');
	let formVenue = $state('');
	let formCoordinator = $state('');

	function openAddActivity() {
		modalMode = 'add';
		editingId = null;
		formName = '';
		formTrack = availableTracks[0]?.name || '';
		userSelectedTrack = false;
		formType = 'Workshop';
		formCredits = creditRules['Workshop'];
		formStatus = 'Active';
		formCategory = 'TECHNICAL';
		formDescription = '';
		formMode = 'Offline';
		const today = new Date().toISOString().split('T')[0];
		formRegDeadline = today;
		formActivityDate = today;
		formVenue = '';
		formCoordinator = '';
		isActivityModalOpen = true;
	}

	function openEditActivity(activity: Activity) {
		modalMode = 'edit';
		editingId = activity.id;
		formName = activity.name;
		formType = activity.type;
		formCredits = activity.credits;
		formStatus = activity.status;
		formCategory = activity.category || 'TECHNICAL';
		formTrack = activity.track || (availableTracks[0]?.name ?? '');
		userSelectedTrack = true;
		formDescription = activity.description || '';
		formMode = activity.mode || 'Offline';
		formRegDeadline = activity.reg_deadline
			? new Date(activity.reg_deadline).toISOString().split('T')[0]
			: '';
		formActivityDate = activity.activity_date
			? new Date(activity.activity_date).toISOString().split('T')[0]
			: '';
		formVenue = activity.venue || '';
		formCoordinator = activity.coordinator || '';
		isActivityModalOpen = true;
	}

	function syncCreditsToType() {
		formCredits = creditRules[formType];
	}

	function syncTrackToCategory() {
		if (userSelectedTrack) return;
		const cat = formCategory.toUpperCase();
		if (cat === 'TECHNICAL' || cat === 'RESEARCH' || cat === 'SPORTS' || cat === 'CULTURAL') {
			if (availableTracks.some((t) => t.name === 'Skill Building')) {
				formTrack = 'Skill Building';
			}
		} else {
			if (availableTracks.some((t) => t.name === 'Personality Development')) {
				formTrack = 'Personality Development';
			}
		}
	}

	async function handleSaveActivity(e: Event) {
		e.preventDefault();
		if (!formName.trim()) return;

		if (
			!formTrack ||
			(availableTracks.length === 0 && !availableTracks.some((t) => t.name === formTrack))
		) {
			triggerToast(
				'Please select a valid track. If no tracks exist, create one in Track Management first.'
			);
			return;
		}

		if (formRegDeadline && formActivityDate && formRegDeadline > formActivityDate) {
			triggerToast('Registration deadline must be on or before the activity date');
			return;
		}

		const token = localStorage.getItem('superadmin_token') || '';
		const payload = {
			name: formName.trim(),
			track: formTrack,
			type: formType,
			credits: formCredits,
			status: formStatus,
			category: formCategory,
			description: formDescription.trim(),
			mode: formMode,
			reg_deadline: formRegDeadline,
			activity_date: formActivityDate,
			venue: formVenue.trim(),
			coordinator: formCoordinator.trim()
		};

		try {
			if (modalMode === 'add') {
				const res = await fetch(`${API_BASE_URL}/api/admin/platform/activities`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`
					},
					body: JSON.stringify(payload)
				});
				if (res.ok) {
					const data = await res.json();
					activities = [data.activity, ...activities];
					triggerToast(`Activity "${payload.name}" published successfully!`);
					isActivityModalOpen = false;
				} else {
					const errorData = await res.json().catch(() => ({}));
					const errMsg = errorData.error || 'Failed to create activity';
					triggerToast(errMsg);
				}
			} else if (editingId) {
				const res = await fetch(`${API_BASE_URL}/api/admin/platform/activities/${editingId}`, {
					method: 'PUT',
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`
					},
					body: JSON.stringify(payload)
				});
				if (res.ok) {
					const data = await res.json();
					activities = activities.map((a) => (a.id === editingId ? data.activity : a));
					triggerToast(`Activity "${payload.name}" updated successfully!`);
					isActivityModalOpen = false;
				} else {
					const errorData = await res.json().catch(() => ({}));
					const errMsg = errorData.error || 'Failed to update activity';
					triggerToast(errMsg);
				}
			}
		} catch (err) {
			console.error(err);
			triggerToast('Network error while saving activity');
		}
	}

	async function handleDeleteActivity(activity: Activity) {
		if (confirm(`Are you sure you want to delete "${activity.name}"?`)) {
			const token = localStorage.getItem('superadmin_token') || '';
			try {
				const res = await fetch(`${API_BASE_URL}/api/admin/platform/activities/${activity.id}`, {
					method: 'DELETE',
					headers: {
						Authorization: `Bearer ${token}`
					}
				});
				if (res.ok) {
					activities = activities.filter((a) => a.id !== activity.id);
					triggerToast(`Activity "${activity.name}" deleted successfully.`);
				} else {
					triggerToast('Failed to delete activity');
				}
			} catch (err) {
				console.error(err);
				triggerToast('Network error while deleting activity');
			}
		}
	}

	async function toggleActivityStatus(activity: Activity) {
		const token = localStorage.getItem('superadmin_token') || '';
		const newStatus = activity.status === 'Active' ? 'Inactive' : 'Active';
		const payload = {
			name: activity.name,
			track: activity.track,
			type: activity.type,
			credits: activity.credits,
			status: newStatus
		};
		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/activities/${activity.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				},
				body: JSON.stringify(payload)
			});
			if (res.ok) {
				const data = await res.json();
				activities = activities.map((a) => (a.id === activity.id ? data.activity : a));
				triggerToast(`"${activity.name}" marked as ${newStatus}.`);
			} else {
				triggerToast('Failed to update status');
			}
		} catch (err) {
			console.error(err);
			triggerToast('Network error while toggling status');
		}
	}

	// ── View Activity Modal ───────────────────────────────────────────────────────
	let isViewModalOpen = $state(false);
	let viewingActivity = $state<Activity | null>(null);

	function openViewActivity(activity: Activity) {
		viewingActivity = activity;
		isViewModalOpen = true;
	}

	// ── Edit Credit Rules Modal ──────────────────────────────────────────────────
	let isCreditModalOpen = $state(false);
	let draftCreditRules = $state<Record<ActivityType, number>>({
		Workshop: 0,
		Seminar: 0,
		Webinar: 0,
		Course: 0
	});

	function openCreditRules() {
		draftCreditRules = { ...creditRules };
		isCreditModalOpen = true;
	}

	function handleSaveCreditRules(e: Event) {
		e.preventDefault();
		creditRules = { ...draftCreditRules };
		triggerToast('Credit rules updated successfully!');
		isCreditModalOpen = false;
	}

	// ── Style Helpers ─────────────────────────────────────────────────────────────
	function trackBadgeClass(track: Track): string {
		return track === 'Personality Development'
			? 'bg-purple-50 text-purple-700 border-purple-100'
			: 'bg-rose-50 text-rose-700 border-rose-100';
	}

	function statusTextClass(status: ActivityStatus): string {
		return status === 'Active' ? 'text-emerald-600' : 'text-slate-400';
	}

	function statusDotClass(status: ActivityStatus): string {
		return status === 'Active' ? 'bg-emerald-600' : 'bg-slate-400';
	}
</script>

<div class="space-y-6">
	<!-- ==================== STAT CARDS ==================== -->
	<section
		class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
		aria-label="Activity metrics overview"
	>
		<!-- Total Activities -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{totalActivitiesCount}</span>
				<div class="p-2.5 rounded-lg bg-slate-100 text-slate-600 border border-slate-200">
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
							d="M3.75 6A2.25 2.25 0 016 3.75h2.25c.597 0 1.142.234 1.543.615l1.8 1.8A2.25 2.25 0 0013.183 6.75H18A2.25 2.25 0 0120.25 9v9A2.25 2.25 0 0118 20.25H6A2.25 2.25 0 013.75 18V6z"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Total activities</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Total campus activities
				</p>
			</div>
		</div>

		<!-- Active Activities -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{activeActivitiesCount}</span>
				<div class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100">
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
							d="M3.75 12h3.75l2.25-6 3.75 12 2.25-6h3.75"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Active activities</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Currently active & open
				</p>
			</div>
		</div>

		<!-- Personality Development -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{personalityDevCount}</span>
				<div class="p-2.5 rounded-lg bg-purple-50 text-purple-600 border border-purple-100">
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
					{personalityDevCount} registered activities
				</p>
			</div>
		</div>

		<!-- Skill Building -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{skillBuildingCount}</span>
				<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100">
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
					{skillBuildingCount} registered activities
				</p>
			</div>
		</div>
	</section>

	<!-- ==================== MAIN GRID: OVERVIEW + MANAGE CREDITS ==================== -->
	<div class="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
		<!-- Activity Management Overview Table (lg:col-span-3) -->
		<div
			class="lg:col-span-3 bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden flex flex-col"
		>
			<!-- Header -->
			<div
				class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-slate-50/20 select-none"
			>
				<div>
					<h3 class="text-base font-bold font-serif text-slate-900">
						Activity Management Overview
					</h3>
					<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
						{filteredActivities.length} of {activities.length} activities
					</p>
				</div>
				<button
					type="button"
					onclick={openAddActivity}
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
					Add Activity
				</button>
			</div>

			<!-- Filters & Search -->
			<div
				class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white select-none"
			>
				<div class="flex flex-wrap gap-1.5">
					{#each filterTabs as tab}
						<button
							type="button"
							onclick={() => setFilter(tab)}
							class="px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all
								{activeFilter === tab
								? 'bg-[#881B1B] text-white shadow-xs'
								: 'bg-slate-50 text-slate-500 hover:bg-slate-100'}"
						>
							{tab}
						</button>
					{/each}
				</div>

				<div class="relative w-full sm:w-64">
					<input
						type="text"
						value={searchQuery}
						oninput={(e) => setSearch((e.target as HTMLInputElement).value)}
						placeholder="Search activity name..."
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

			<!-- Table View -->
			<div class="overflow-x-auto flex-grow no-scrollbar">
				<table class="w-full text-left border-collapse min-w-[600px]">
					<thead>
						<tr
							class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
						>
							<th class="py-3.5 px-5">Activity Name</th>
							<th class="py-3.5 px-5">Track</th>
							<th class="py-3.5 px-5">Type</th>
							<th class="py-3.5 px-5">Credits</th>
							<th class="py-3.5 px-5">Status</th>
							<th class="py-3.5 px-5 text-center">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 text-xs font-sans">
						{#if paginatedActivities.length === 0}
							<tr>
								<td colspan="6" class="py-8 text-center text-slate-400 font-semibold select-none">
									No activities found matching search filters.
								</td>
							</tr>
						{:else}
							{#each paginatedActivities as activity (activity.id)}
								<tr class="hover:bg-slate-50/30 transition-colors">
									<td class="py-4 px-5">
										<div class="flex flex-col">
											<span class="font-bold text-slate-800">{activity.name}</span>
											<span class="text-[10px] text-slate-400 font-semibold mt-0.5 select-all"
												>{activity.id}</span
											>
										</div>
									</td>
									<td class="py-4 px-5">
										<span
											class="inline-flex items-center px-2 py-0.5 text-[10px] font-bold rounded-full border {trackBadgeClass(
												activity.track
											)}"
										>
											{activity.track || 'Unassigned'}
										</span>
									</td>
									<td class="py-4 px-5 text-slate-500 font-semibold">{activity.type}</td>
									<td class="py-4 px-5 font-bold text-slate-800">
										{activity.credits}
									</td>
									<td class="py-4 px-5">
										<button
											type="button"
											onclick={() => toggleActivityStatus(activity)}
											class="inline-flex items-center gap-1.5 font-bold {statusTextClass(
												activity.status
											)} hover:opacity-70 transition-opacity"
										>
											<span
												class="w-1.5 h-1.5 rounded-full shrink-0 {statusDotClass(activity.status)}"
											></span>
											{activity.status}
										</button>
									</td>
									<td class="py-4 px-5">
										<div class="flex items-center justify-center gap-3 text-slate-400">
											<button
												type="button"
												onclick={() => openViewActivity(activity)}
												aria-label="View activity"
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
												onclick={() => openEditActivity(activity)}
												aria-label="Edit activity"
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
														d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
													/>
												</svg>
											</button>
											<button
												type="button"
												onclick={() => handleDeleteActivity(activity)}
												aria-label="Delete activity"
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

			<!-- Footer / Pagination -->
			<div
				class="p-4 border-t border-slate-100 bg-slate-50/30 flex items-center justify-between gap-3 select-none"
			>
				<span class="text-slate-500 font-semibold text-[11px]">
					Showing {paginatedActivities.length} of {filteredActivities.length} activities
				</span>
				{#if totalPages > 1}
					<div class="flex items-center gap-1.5">
						{#each Array(totalPages) as _, i}
							<button
								type="button"
								onclick={() => goToPage(i + 1)}
								class="w-7 h-7 flex items-center justify-center rounded-lg text-[11px] font-bold transition-colors
									{currentPage === i + 1
									? 'bg-[#881B1B] text-white'
									: 'bg-white border border-slate-200 text-slate-500 hover:bg-slate-50'}"
							>
								{i + 1}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Manage Credits Panel (lg:col-span-1) -->
		<div class="bg-white border border-slate-200 rounded-xl p-5 shadow-xs space-y-4">
			<div class="flex items-center gap-3">
				<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100 shrink-0">
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
							d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497"
						/>
					</svg>
				</div>
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">Manage Credits</h3>
					<p class="text-[10px] font-semibold text-slate-400 mt-0.5">Configure credit values</p>
				</div>
			</div>
			<div class="h-px bg-slate-100"></div>

			<div class="space-y-1">
				{#each Object.keys(creditRules) as ActivityType[] as type}
					<div
						class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0"
					>
						<span class="text-xs font-semibold text-slate-600">{type}</span>
						<span class="text-xs font-bold text-slate-800">
							{creditRules[type]} cr
						</span>
					</div>
				{/each}
			</div>

			<button
				type="button"
				onclick={openCreditRules}
				class="w-full mt-2 inline-flex items-center justify-center gap-1.5 py-2.5 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
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
						d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
					/>
				</svg>
				Edit Credit Rules
			</button>
		</div>
	</div>
</div>

<!-- ==================== TOAST NOTIFICATION CONTAINER ==================== -->
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

<!-- ==================== ADD / EDIT ACTIVITY MODAL ==================== -->
{#if isActivityModalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isActivityModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<form
			onsubmit={handleSaveActivity}
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans max-h-[90vh] overflow-y-auto"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">
						{modalMode === 'add' ? 'Publish New Activity' : 'Edit Activity'}
					</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						{modalMode === 'add' ? 'Register Extracurricular Activity' : 'Update activity details'}
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isActivityModalOpen = false)}
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
					<label for="act-name" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
						>ACTIVITY NAME *</label
					>
					<input
						id="act-name"
						type="text"
						bind:value={formName}
						placeholder="e.g. Cyber Security Summit"
						required
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<label for="act-desc" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
						>DESCRIPTION</label
					>
					<textarea
						id="act-desc"
						bind:value={formDescription}
						placeholder="Provide a detailed description of the activity..."
						rows="3"
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 resize-none"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label for="act-track" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>TRACK *</label
						>
						<select
							id="act-track"
							bind:value={formTrack}
							onchange={() => (userSelectedTrack = true)}
							disabled={availableTracks.length === 0 && !formTrack}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355 disabled:bg-slate-100 disabled:text-slate-400"
						>
							{#if availableTracks.length === 0 && !formTrack}
								<option value="" disabled selected
									>No tracks available. Please create a track first.</option
								>
							{:else}
								{#each availableTracks as trackOption}
									<option value={trackOption.name}>{trackOption.name}</option>
								{/each}
								{#if !availableTracks.some((t) => t.name === formTrack) && formTrack}
									<option value={formTrack}>{formTrack}</option>
								{/if}
							{/if}
						</select>
						{#if availableTracks.length === 0 && !formTrack}
							<p class="text-[11px] text-amber-600 font-medium mt-1">
								No active tracks exist in the database. Please add a track in Track Management
								first.
							</p>
						{/if}
					</div>

					<div class="flex flex-col gap-1.5">
						<label for="act-type" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>TYPE *</label
						>
						<select
							id="act-type"
							bind:value={formType}
							onchange={syncCreditsToType}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="Workshop">Workshop</option>
							<option value="Seminar">Seminar</option>
							<option value="Webinar">Webinar</option>
							<option value="Course">Course</option>
						</select>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label
							for="act-category"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider">CATEGORY *</label
						>
						<select
							id="act-category"
							bind:value={formCategory}
							onchange={syncTrackToCategory}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="TECHNICAL">Technical</option>
							<option value="SPORTS">Sports</option>
							<option value="CULTURAL">Cultural</option>
							<option value="SOCIAL SERVICE">Social Service</option>
							<option value="RESEARCH">Research</option>
							<option value="PUBLIC SPEAKING">Public Speaking</option>
							<option value="LEADERSHIP">Leadership</option>
						</select>
					</div>

					<div class="flex flex-col gap-1.5">
						<label for="act-mode" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>MODE *</label
						>
						<select
							id="act-mode"
							bind:value={formMode}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="Offline">Offline</option>
							<option value="Online">Online</option>
							<option value="Hybrid">Hybrid</option>
						</select>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label
							for="act-reg-deadline"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>REGISTRATION DEADLINE *</label
						>
						<input
							id="act-reg-deadline"
							type="date"
							bind:value={formRegDeadline}
							required
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<label for="act-date" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>ACTIVITY DATE *</label
						>
						<input
							id="act-date"
							type="date"
							bind:value={formActivityDate}
							required
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label for="act-venue" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>VENUE</label
						>
						<input
							id="act-venue"
							type="text"
							bind:value={formVenue}
							placeholder="e.g. IIPS Main Ground"
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<label
							for="act-coordinator"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider">COORDINATOR</label
						>
						<input
							id="act-coordinator"
							type="text"
							bind:value={formCoordinator}
							placeholder="e.g. NSS Cell"
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label
							for="act-credits"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider">CREDITS *</label
						>
						<input
							id="act-credits"
							type="number"
							bind:value={formCredits}
							min="1"
							max="50"
							required
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>

					<div class="flex flex-col gap-1.5">
						<label for="act-status" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>STATUS</label
						>
						<select
							id="act-status"
							bind:value={formStatus}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="Active">Active</option>
							<option value="Inactive">Inactive</option>
						</select>
					</div>
				</div>
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
			>
				<button
					type="button"
					onclick={() => (isActivityModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Cancel
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					{modalMode === 'add' ? 'Publish Activity' : 'Save Changes'}
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- ==================== VIEW ACTIVITY MODAL ==================== -->
{#if isViewModalOpen && viewingActivity}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isViewModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<div
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans max-h-[90vh]"
		>
			<!-- Header -->
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">View Activity</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Activity Details ({viewingActivity.id})
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isViewModalOpen = false)}
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

			<!-- Body - Scrollable Form Structure -->
			<div class="p-6 space-y-4 overflow-y-auto flex-grow">
				<!-- Activity Name -->
				<div class="flex flex-col gap-1.5">
					<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
						ACTIVITY NAME
					</span>
					<div
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
					>
						{viewingActivity.name}
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
						{viewingActivity.description || 'No description provided.'}
					</div>
				</div>

				<!-- Track & Type -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							TRACK
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.track || 'Unassigned'}
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							TYPE
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.type}
						</div>
					</div>
				</div>

				<!-- Category & Mode -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							CATEGORY
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.category || 'TECHNICAL'}
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							MODE
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.mode || 'Offline'}
						</div>
					</div>
				</div>

				<!-- Registration Deadline & Activity Date -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							REGISTRATION DEADLINE
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.reg_deadline
								? new Date(viewingActivity.reg_deadline).toISOString().split('T')[0]
								: 'N/A'}
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							ACTIVITY DATE
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.activity_date
								? new Date(viewingActivity.activity_date).toISOString().split('T')[0]
								: 'N/A'}
						</div>
					</div>
				</div>

				<!-- Venue & Coordinator -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							VENUE
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.venue || 'N/A'}
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							COORDINATOR
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.coordinator || 'N/A'}
						</div>
					</div>
				</div>

				<!-- Credits & Status -->
				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							CREDITS
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-medium text-slate-800 bg-slate-50/50"
						>
							{viewingActivity.credits} cr
						</div>
					</div>

					<div class="flex flex-col gap-1.5">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider uppercase">
							STATUS
						</span>
						<div
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-bold bg-slate-50/50 flex items-center gap-1.5 {statusTextClass(
								viewingActivity.status
							)}"
						>
							<span
								class="w-1.5 h-1.5 rounded-full shrink-0 {statusDotClass(viewingActivity.status)}"
							></span>
							{viewingActivity.status}
						</div>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30 shrink-0"
			>
				<button
					type="button"
					onclick={() => (isViewModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Close
				</button>
				<button
					type="button"
					onclick={() => {
						isViewModalOpen = false;
						if (viewingActivity) openEditActivity(viewingActivity);
					}}
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Edit Activity
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== EDIT CREDIT RULES MODAL ==================== -->
{#if isCreditModalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isCreditModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<form
			onsubmit={handleSaveCreditRules}
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">Edit Credit Rules</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Configure credit values by activity type
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isCreditModalOpen = false)}
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
				{#each Object.keys(draftCreditRules) as ActivityType[] as type}
					<div class="flex items-center justify-between gap-4">
						<label for="credit-{type}" class="text-xs font-bold text-slate-700">{type}</label>
						<input
							id="credit-{type}"
							type="number"
							bind:value={draftCreditRules[type]}
							min="1"
							max="50"
							class="w-24 px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 text-right focus:outline-none focus:border-slate-355"
						/>
					</div>
				{/each}
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
			>
				<button
					type="button"
					onclick={() => (isCreditModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Cancel
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Save Credit Rules
				</button>
			</div>
		</form>
	</div>
{/if}

<style>
	/* Hide the horizontal scrollbar on the activity table while keeping it scrollable */
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
