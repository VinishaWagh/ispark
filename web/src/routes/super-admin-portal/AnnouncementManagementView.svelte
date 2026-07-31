<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { API_BASE_URL } from '$lib/config';

	// ── Types ────────────────────────────────────────────────────────────────
	type AnnouncementStatus = 'active' | 'draft' | 'scheduled' | 'expired';
	type AudienceType = 'Students' | 'Mentors' | 'All Users';
	type Category = 'General' | 'Academic' | 'Activities' | 'Events';
	type Priority = 'Low' | 'Medium' | 'High';

	interface Announcement {
		id: number;
		title: string;
		description: string;
		category: Category;
		audience: AudienceType;
		priority: Priority;
		publishDate: string;
		expiryDate: string;
		status: AnnouncementStatus;
	}

	interface AnnouncementApiResponse {
		id: number;
		title: string;
		description: string;
		category: Category;
		audience: AudienceType;
		priority: Priority;
		publish_date: string;
		expiry_date: string;
		status: AnnouncementStatus;
	}

	type FilterKey = 'All' | 'Active' | 'Draft' | 'Scheduled' | 'Expired';

	const announcementsEndpoint = `${API_BASE_URL}/api/admin/platform/announcements`;
	let announcements = $state<Announcement[]>([]);
	let isLoading = $state(true);
	let loadError = $state('');
	let isSaving = $state(false);
	let actionId = $state<number | null>(null);

	function fromApiAnnouncement(item: AnnouncementApiResponse): Announcement {
		return {
			id: item.id,
			title: item.title,
			description: item.description,
			category: item.category,
			audience: item.audience,
			priority: item.priority,
			publishDate: item.publish_date,
			expiryDate: item.expiry_date,
			status: item.status
		};
	}

	async function apiRequest<T>(
		url: string,
		options: { method?: string; body?: string; headers?: Record<string, string> } = {}
	): Promise<T> {
		const token = localStorage.getItem('superadmin_token') || '';
		const response = await fetch(url, {
			...options,
			headers: {
				Authorization: `Bearer ${token}`,
				...(options.body ? { 'Content-Type': 'application/json' } : {}),
				...options.headers
			}
		});
		const body = await response.json().catch(() => ({}));
		if (!response.ok) {
			throw new Error(body.error || 'Announcement request failed');
		}
		return body as T;
	}

	async function loadAnnouncements() {
		isLoading = true;
		loadError = '';
		try {
			const response = await apiRequest<{ announcements: AnnouncementApiResponse[] }>(
				announcementsEndpoint
			);
			announcements = response.announcements.map(fromApiAnnouncement);
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'Could not load announcements.';
		} finally {
			isLoading = false;
		}
	}

	onMount(loadAnnouncements);

	let announcementFilter = $state<FilterKey>('All');
	let announcementSearch = $state('');

	let filteredAnnouncements = $derived(
		announcements.filter((a) => {
			const matchesFilter =
				announcementFilter === 'All' || a.status === announcementFilter.toLowerCase();
			const query = announcementSearch.toLowerCase().trim();
			const matchesSearch =
				query === '' ||
				a.title.toLowerCase().includes(query) ||
				a.category.toLowerCase().includes(query) ||
				a.audience.toLowerCase().includes(query);
			return matchesFilter && matchesSearch;
		})
	);

	// ── Stat card derivations ────────────────────────────────────────────────
	let totalAnnouncementsCount = $derived(announcements.length);
	let activeAnnouncementsCount = $derived(
		announcements.filter((a) => a.status === 'active').length
	);
	let scheduledAnnouncementsCount = $derived(
		announcements.filter((a) => a.status === 'scheduled').length
	);
	let expiredAnnouncementsCount = $derived(
		announcements.filter((a) => a.status === 'expired').length
	);

	const audienceStyles: Record<AudienceType, string> = {
		Students: 'bg-blue-50 text-blue-700',
		Mentors: 'bg-violet-50 text-violet-700',
		'All Users': 'bg-slate-100 text-slate-700'
	};

	const priorityStyles: Record<Priority, string> = {
		Low: 'bg-slate-50 text-slate-600 border-slate-200',
		Medium: 'bg-amber-50 text-amber-700 border-amber-100',
		High: 'bg-red-50 text-red-600 border-red-100'
	};

	const priorityDotStyles: Record<Priority, string> = {
		Low: 'bg-slate-400',
		Medium: 'bg-amber-500',
		High: 'bg-red-500'
	};

	function formatDisplayDate(iso: string): string {
		if (!iso) return '';
		const d = new Date(`${iso}T00:00:00`);
		if (Number.isNaN(d.getTime())) return iso;
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function announcementStatusClass(status: AnnouncementStatus): string {
		switch (status) {
			case 'active':
				return 'bg-emerald-50 text-emerald-700 border-emerald-100';
			case 'scheduled':
				return 'bg-amber-50 text-amber-700 border-amber-100';
			case 'expired':
				return 'bg-red-50 text-red-600 border-red-100';
			default:
				return 'bg-slate-100 text-slate-500 border-slate-200';
		}
	}
	function announcementStatusDot(status: AnnouncementStatus): string {
		switch (status) {
			case 'active':
				return 'bg-emerald-500';
			case 'scheduled':
				return 'bg-amber-500';
			case 'expired':
				return 'bg-red-500';
			default:
				return 'bg-slate-400';
		}
	}

	function announcementStatusLabel(status: AnnouncementStatus): string {
		return status.charAt(0).toUpperCase() + status.slice(1);
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

	// ── Create / Edit Announcement modal ─────────────────────────────────────
	let isFormModalOpen = $state(false);
	let formMode = $state<'create' | 'edit'>('create');
	let editingId = $state<number | null>(null);

	let formTitle = $state('');
	let formDescription = $state('');
	let formAudience = $state<AudienceType>('Students');
	let formCategory = $state<Category>('General');
	let formPriority = $state<Priority>('Medium');
	let formPublishDate = $state('');
	let formExpiryDate = $state('');
	let formStatus = $state<AnnouncementStatus>('draft');
	let formError = $state('');

	function resetForm() {
		formTitle = '';
		formDescription = '';
		formCategory = 'General';
		formAudience = 'Students';
		formPriority = 'Medium';
		formPublishDate = '';
		formExpiryDate = '';
		formStatus = 'draft';
		formError = '';
		editingId = null;
	}

	function openCreateAnnouncement() {
		resetForm();
		formMode = 'create';
		const now = Date.now();
		formPublishDate = new Date(now).toISOString().slice(0, 10);
		formExpiryDate = new Date(now + 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
		isFormModalOpen = true;
	}

	function openEditAnnouncement(item: Announcement) {
		formMode = 'edit';
		editingId = item.id;
		formTitle = item.title;
		formDescription = item.description;
		formCategory = item.category;
		formAudience = item.audience;
		formPriority = item.priority;
		formPublishDate = item.publishDate;
		formExpiryDate = item.expiryDate;
		formStatus = item.status;
		formError = '';
		isFormModalOpen = true;
	}

	async function handleSaveAnnouncement(e: Event) {
		e.preventDefault();

		if (!formTitle.trim() || !formPublishDate || !formExpiryDate) {
			formError = 'Please fill in all required fields.';
			return;
		}

		if (new Date(formExpiryDate) < new Date(formPublishDate)) {
			formError = 'Expiry date cannot be before the publish date.';
			return;
		}

		isSaving = true;
		formError = '';
		try {
			const response = await apiRequest<{ announcement: AnnouncementApiResponse }>(
				formMode === 'create' ? announcementsEndpoint : `${announcementsEndpoint}/${editingId}`,
				{
					method: formMode === 'create' ? 'POST' : 'PUT',
					body: JSON.stringify({
						title: formTitle.trim(),
						description: formDescription.trim(),
						category: formCategory,
						audience: formAudience,
						priority: formPriority,
						publish_date: formPublishDate,
						expiry_date: formExpiryDate,
						status: formStatus
					})
				}
			);
			const saved = fromApiAnnouncement(response.announcement);
			if (formMode === 'create') {
				announcements = [saved, ...announcements];
			} else {
				announcements = announcements.map((item) => (item.id === saved.id ? saved : item));
			}
			const action =
				saved.status === 'draft'
					? 'saved as a draft'
					: saved.status === 'scheduled'
						? 'scheduled successfully'
						: saved.status === 'active'
							? 'published successfully'
							: 'updated successfully';
			triggerToast(`Announcement "${saved.title}" ${action}.`);
			isFormModalOpen = false;
			resetForm();
		} catch (error) {
			formError = error instanceof Error ? error.message : 'Could not save the announcement.';
		} finally {
			isSaving = false;
		}
	}

	// ── View Announcement modal ──────────────────────────────────────────────
	let isViewAnnouncementModalOpen = $state(false);
	let viewAnnouncement = $state<Announcement | null>(null);

	function openViewAnnouncement(item: Announcement) {
		viewAnnouncement = item;
		isViewAnnouncementModalOpen = true;
	}

	// ── Delete ────────────────────────────────────────────────────────────────
	async function handleDeleteAnnouncement(item: Announcement) {
		if (!confirm(`Are you sure you want to delete "${item.title}"?`)) return;
		actionId = item.id;
		loadError = '';
		try {
			await apiRequest(`${announcementsEndpoint}/${item.id}`, { method: 'DELETE' });
			announcements = announcements.filter((announcement) => announcement.id !== item.id);
			if (viewAnnouncement?.id === item.id) {
				isViewAnnouncementModalOpen = false;
				viewAnnouncement = null;
			}
			triggerToast(`Announcement "${item.title}" removed successfully.`);
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'Could not delete the announcement.';
		} finally {
			actionId = null;
		}
	}

	// ── Publish (draft → active) ─────────────────────────────────────────────
	async function handlePublishAnnouncement(item: Announcement) {
		actionId = item.id;
		loadError = '';
		try {
			const response = await apiRequest<{ announcement: AnnouncementApiResponse }>(
				`${announcementsEndpoint}/${item.id}/publish`,
				{ method: 'POST' }
			);
			const published = fromApiAnnouncement(response.announcement);
			announcements = announcements.map((announcement) =>
				announcement.id === published.id ? published : announcement
			);
			if (viewAnnouncement?.id === published.id) viewAnnouncement = published;
			triggerToast(`Announcement "${published.title}" is now live.`);
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'Could not publish the announcement.';
		} finally {
			actionId = null;
		}
	}
</script>

<!-- ==================== ANNOUNCEMENT MANAGEMENT ==================== -->
<div class="space-y-6">
	<!-- Overview Stat Cards -->
	<section
		class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
		aria-label="Announcement management overview"
	>
		<!-- Total Announcements -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{totalAnnouncementsCount}</span>
				<div class="p-2.5 rounded-lg bg-slate-100 text-slate-600 border border-slate-200">
					<!-- Megaphone icon -->
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
							d="M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
					Total Announcements
				</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					All-time records
				</p>
			</div>
		</div>

		<!-- Active -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{activeAnnouncementsCount}</span>
				<div class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100">
					<!-- Check circle icon -->
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
							d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Active</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Currently visible
				</p>
			</div>
		</div>

		<!-- Scheduled -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900"
					>{scheduledAnnouncementsCount}</span
				>
				<div class="p-2.5 rounded-lg bg-purple-50 text-purple-600 border border-purple-100">
					<!-- Pencil icon -->
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
							d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931z"
						/>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M19.5 13.5v4.875c0 .621-.504 1.125-1.125 1.125H5.625a1.125 1.125 0 01-1.125-1.125V6.75c0-.621.504-1.125 1.125-1.125h4.875"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Scheduled</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Queued for publishing
				</p>
			</div>
		</div>

		<!-- Expired -->
		<div
			class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
		>
			<div class="flex items-center justify-between">
				<span class="text-2xl font-bold font-serif text-slate-900">{expiredAnnouncementsCount}</span
				>
				<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100">
					<!-- Warning triangle icon -->
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
							d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
						/>
					</svg>
				</div>
			</div>
			<div class="mt-4">
				<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Expired</h3>
				<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
					Past expiry date
				</p>
			</div>
		</div>
	</section>

	<!-- Announcement Management Overview -->
	<section class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
		<!-- Header -->
		<div
			class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-50/20 select-none"
		>
			<div>
				<h3 class="text-sm font-bold font-serif text-slate-905">
					Announcement Management Overview
				</h3>
				<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
					{filteredAnnouncements.length} of {announcements.length} announcements
				</p>
			</div>

			<button
				type="button"
				onclick={openCreateAnnouncement}
				class="inline-flex items-center justify-center gap-1.5 w-full sm:w-auto px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none shrink-0"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="2"
					stroke="currentColor"
					class="w-4 h-4"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Create Announcement
			</button>
		</div>

		<!-- Filters & Search -->
		<div
			class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white select-none"
		>
			<div class="flex flex-wrap gap-1.5">
				{#each ['All', 'Active', 'Draft', 'Scheduled', 'Expired'] as filterOption}
					<button
						type="button"
						onclick={() => (announcementFilter = filterOption as FilterKey)}
						class="px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all
							{announcementFilter === filterOption
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
					bind:value={announcementSearch}
					placeholder="Search title, category, audience"
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

		{#if loadError}
			<div
				class="px-5 py-3 bg-red-50 border-b border-red-100 text-xs font-semibold text-red-700"
				role="alert"
			>
				{loadError}
				<button type="button" onclick={loadAnnouncements} class="ml-2 font-bold underline"
					>Retry</button
				>
			</div>
		{/if}

		<!-- Table -->
		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr
						class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
					>
						<th class="py-3.5 px-5">Announcement Title</th>
						<th class="py-3.5 px-5">Category</th>
						<th class="py-3.5 px-5">Priority</th>
						<th class="py-3.5 px-5">Target Audience</th>
						<th class="py-3.5 px-5">Publish Date</th>
						<th class="py-3.5 px-5">Expiry Date</th>
						<th class="py-3.5 px-5">Status</th>
						<th class="py-3.5 px-5">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100 text-xs font-sans">
					{#if isLoading}
						<tr>
							<td colspan="8" class="py-8 text-center text-slate-400 font-semibold select-none">
								Loading announcements...
							</td>
						</tr>
					{:else if filteredAnnouncements.length === 0}
						<tr>
							<td colspan="8" class="py-8 text-center text-slate-400 font-semibold select-none">
								No announcements found matching search filters.
							</td>
						</tr>
					{:else}
						{#each filteredAnnouncements as item (item.id)}
							<tr class="hover:bg-slate-50/30 transition-colors">
								<td class="py-4 px-5 font-bold text-slate-800 align-top max-w-sm">
									{item.title}
								</td>
								<td class="py-4 px-5 text-slate-600 font-semibold align-top">
									{item.category}
								</td>
								<td class="py-4 px-5 align-top">
									<span
										class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border {priorityStyles[
											item.priority
										]}"
									>
										<span class="w-1.5 h-1.5 rounded-full {priorityDotStyles[item.priority]}"
										></span>
										{item.priority}
									</span>
								</td>
								<td class="py-4 px-5 align-top">
									<span
										class="inline-block px-2.5 py-1 rounded-md text-[11px] font-semibold {audienceStyles[
											item.audience
										]}"
									>
										{item.audience}
									</span>
								</td>
								<td class="py-4 px-5 text-slate-500 font-semibold align-top whitespace-nowrap">
									{formatDisplayDate(item.publishDate)}
								</td>
								<td class="py-4 px-5 text-slate-500 font-semibold align-top whitespace-nowrap">
									{formatDisplayDate(item.expiryDate)}
								</td>
								<td class="py-4 px-5 align-top">
									<span
										class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border {announcementStatusClass(
											item.status
										)}"
									>
										<span class="w-1.5 h-1.5 rounded-full {announcementStatusDot(item.status)}"
										></span>
										{announcementStatusLabel(item.status)}
									</span>
								</td>
								<td class="py-4 px-5 align-top">
									<div class="flex items-center gap-3">
										{#if item.status === 'draft'}
											<button
												type="button"
												onclick={() => handlePublishAnnouncement(item)}
												disabled={actionId === item.id}
												aria-label="Publish announcement"
												title="Publish"
												class="text-emerald-500 hover:text-emerald-700 transition-colors p-0.5"
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
														d="M4.5 12.75l6 6 9-13.5"
													/>
												</svg>
											</button>
										{/if}
										<button
											type="button"
											onclick={() => openViewAnnouncement(item)}
											aria-label="View announcement"
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
											onclick={() => openEditAnnouncement(item)}
											disabled={actionId === item.id}
											aria-label="Edit announcement"
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
											onclick={() => handleDeleteAnnouncement(item)}
											disabled={actionId === item.id}
											aria-label="Delete announcement"
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
			<span>Showing {filteredAnnouncements.length} of {announcements.length} announcements</span>
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

<!-- Create / Edit Announcement Modal -->
{#if isFormModalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isFormModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<form
			onsubmit={handleSaveAnnouncement}
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans max-h-[90vh]"
		>
			<div
				class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30 shrink-0"
			>
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">
						{formMode === 'create' ? 'Create New Announcement' : 'Edit Announcement'}
					</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Platform-wide notice
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isFormModalOpen = false)}
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

			<div class="p-6 space-y-4 overflow-y-auto">
				<div class="flex flex-col gap-1.5">
					<label for="ann-title" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
						>TITLE *</label
					>
					<input
						id="ann-title"
						type="text"
						bind:value={formTitle}
						placeholder="e.g. Mid-Semester Activity Submission Deadline"
						required
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
					/>
				</div>

				<div class="flex flex-col gap-1.5">
					<label
						for="ann-description"
						class="text-[10px] font-extrabold text-slate-650 tracking-wider">DESCRIPTION</label
					>
					<textarea
						id="ann-description"
						bind:value={formDescription}
						rows="3"
						placeholder="Brief details for students or faculty..."
						class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 resize-none"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label
							for="ann-category"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider"
						>
							CATEGORY *
						</label>

						<select
							id="ann-category"
							bind:value={formCategory}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="General">General</option>
							<option value="Academic">Academic</option>
							<option value="Activities">Activities</option>
							<option value="Events">Events</option>
						</select>
					</div>
					<div class="flex flex-col gap-1.5">
						<label
							for="ann-audience"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>TARGET AUDIENCE *</label
						>
						<select
							id="ann-audience"
							bind:value={formAudience}
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 bg-white focus:outline-none focus:border-slate-355"
						>
							<option value="Students">Students</option>
							<option value="Mentors">Mentors</option>
							<option value="All Users">All Users</option>
						</select>
					</div>
				</div>

				<div class="flex flex-col gap-1.5">
					<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">PRIORITY *</span>
					<div class="grid grid-cols-3 gap-2">
						{#each ['Low', 'Medium', 'High'] as const as p}
							<button
								type="button"
								onclick={() => (formPriority = p)}
								class="px-3 py-2 rounded-lg text-xs font-bold border transition-all
								{formPriority === p
									? 'bg-[#881B1B]/10 text-[#881B1B] border-[#881B1B]/30'
									: 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50'}"
							>
								{p}
							</button>
						{/each}
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1.5">
						<label
							for="ann-publish"
							class="text-[10px] font-extrabold text-slate-650 tracking-wider">PUBLISH DATE *</label
						>
						<input
							id="ann-publish"
							type="date"
							bind:value={formPublishDate}
							required
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>
					<div class="flex flex-col gap-1.5">
						<label for="ann-expiry" class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>EXPIRY DATE *</label
						>
						<input
							id="ann-expiry"
							type="date"
							bind:value={formExpiryDate}
							required
							class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
						/>
					</div>
				</div>

				<div class="flex flex-col gap-1.5">
					<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">STATUS</span>
					<div class="grid grid-cols-3 gap-2">
						{#each ['draft', 'scheduled', 'active', 'expired'] as const as s}
							<button
								type="button"
								onclick={() => (formStatus = s)}
								class="px-3 py-2 rounded-lg text-xs font-bold border transition-all capitalize
								{formStatus === s
									? 'bg-[#881B1B]/10 text-[#881B1B] border-[#881B1B]/30'
									: 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50'}"
							>
								{s}
							</button>
						{/each}
					</div>
				</div>

				{#if formError}
					<div
						class="p-3 bg-red-50 border border-red-200 rounded-lg text-[11px] font-semibold text-red-700"
						role="alert"
					>
						{formError}
					</div>
				{/if}
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30 shrink-0"
			>
				<button
					type="button"
					onclick={() => (isFormModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSaving}
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					{isSaving ? 'Saving...' : formMode === 'create' ? 'Save Announcement' : 'Save Changes'}
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- View Announcement Modal -->
{#if isViewAnnouncementModalOpen && viewAnnouncement}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		onclick={(e) => {
			if (e.target === e.currentTarget) isViewAnnouncementModalOpen = false;
		}}
		transition:fade={{ duration: 150 }}
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
	>
		<div
			class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
		>
			<div class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30">
				<div>
					<h3 class="text-sm font-bold font-serif text-slate-900">{viewAnnouncement.title}</h3>
					<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
						Announcement Details
					</p>
				</div>
				<button
					type="button"
					onclick={() => (isViewAnnouncementModalOpen = false)}
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

			<div class="p-6 space-y-4 text-xs font-sans">
				{#if viewAnnouncement.description}
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">DESCRIPTION</span
						>
						<p class="text-slate-700 font-semibold leading-relaxed">
							{viewAnnouncement.description}
						</p>
					</div>
				{/if}

				<div class="grid grid-cols-3 gap-4 pt-2">
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>TARGET AUDIENCE</span
						>
						<span
							class="inline-block px-2.5 py-1 w-fit rounded-md text-[11px] font-semibold {audienceStyles[
								viewAnnouncement.audience
							]}"
						>
							{viewAnnouncement.audience}
						</span>
					</div>
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">PRIORITY</span>
						<span
							class="inline-flex items-center gap-1.5 px-2 py-0.5 w-fit text-[10px] font-bold uppercase rounded-full border {priorityStyles[
								viewAnnouncement.priority
							]}"
						>
							<span class="w-1.5 h-1.5 rounded-full {priorityDotStyles[viewAnnouncement.priority]}"
							></span>
							{viewAnnouncement.priority}
						</span>
					</div>
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">STATUS</span>
						<span
							class="inline-flex items-center gap-1.5 px-2 py-0.5 w-fit text-[10px] font-bold uppercase rounded-full border {announcementStatusClass(
								viewAnnouncement.status
							)}"
						>
							<span
								class="w-1.5 h-1.5 rounded-full {announcementStatusDot(viewAnnouncement.status)}"
							></span>
							{announcementStatusLabel(viewAnnouncement.status)}
						</span>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider"
							>PUBLISH DATE</span
						>
						<span class="text-sm font-bold text-slate-900"
							>{formatDisplayDate(viewAnnouncement.publishDate)}</span
						>
					</div>
					<div class="flex flex-col gap-1">
						<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">EXPIRY DATE</span
						>
						<span class="text-sm font-bold text-slate-900"
							>{formatDisplayDate(viewAnnouncement.expiryDate)}</span
						>
					</div>
				</div>
			</div>

			<div
				class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
			>
				<button
					type="button"
					onclick={() => (isViewAnnouncementModalOpen = false)}
					class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Close
				</button>
				{#if viewAnnouncement.status === 'draft'}
					<button
						type="button"
						onclick={() => viewAnnouncement && handlePublishAnnouncement(viewAnnouncement)}
						disabled={actionId === viewAnnouncement.id}
						class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
					>
						Publish
					</button>
				{/if}
				<button
					type="button"
					onclick={() => {
						isViewAnnouncementModalOpen = false;
						if (viewAnnouncement) openEditAnnouncement(viewAnnouncement);
					}}
					class="px-4 py-2 bg-[#881B1B] hover:bg-[#721616] text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
				>
					Edit Announcement
				</button>
			</div>
		</div>
	</div>
{/if}
