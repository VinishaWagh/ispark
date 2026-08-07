<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fade, slide } from 'svelte/transition';
	import { API_BASE_URL } from '$lib/config';
	import AdminDashboardView from './AdminDashboardView.svelte';
	import AdminPlaceholderView from './AdminPlaceholderView.svelte';
	import ActivityMonitoringView from './ActivityMonitoringView.svelte';
	import BatchAnalyticsView from './BatchAnalyticsView.svelte';
	import AdminStudentManagementView from './AdminStudentManagementView.svelte';
	import AdminCertificateVerificationView from './AdminCertificateVerificationView.svelte';
	import AdminProfileView from './AdminProfileView.svelte';
	import AdminEditProfileView from './AdminEditProfileView.svelte';

	interface AdminProfile {
		admin_id: string;
		name: string;
		email: string;
		role: string;
		assigned_batch: string;
		created_at: string;
		updated_at: string;
	}

	let admin = $state<AdminProfile | null>(null);
	let stats = $state<{
		assigned_students: number;
		verified_certificates: number;
		pending_reviews: number;
		supervised_activities: number;
	} | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		const token = localStorage.getItem('admin_token');
		if (!token) {
			goto('/admin-portal');
			return;
		}
		if (localStorage.getItem('admin_must_change_password') === 'true') {
			goto('/admin-portal/update');
			return;
		}

		try {
			const response = await fetch(`${API_BASE_URL}/api/admin/profile`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to fetch admin profile');
			}

			const data = await response.json();
			admin = data.admin;
			stats = data.stats;
		} catch (err) {
			console.error(err);
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	function getInitials(name: string): string {
		if (!name) return 'A';
		const parts = name.split(' ').filter((part) => {
			const lower = part.toLowerCase();
			return (
				lower !== 'dr.' &&
				lower !== 'prof.' &&
				lower !== 'mr.' &&
				lower !== 'ms.' &&
				lower !== 'mrs.'
			);
		});
		if (parts.length === 0) return 'A';
		if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	// Sidebar menu items list for Admin Portal
	const menuItems = [
		{
			name: 'Dashboard',
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			name: 'Student Management',
			icon: 'M17 20h5v-2a4 4 0 00-5.356-3.772M9 20H4v-2a4 4 0 015.356-3.772M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z'
		},
		{
			name: 'Certificate Verification',
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			name: 'Activity Monitoring',
			icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		},
		{
			name: 'Batch Analytics',
			icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
		},
		{
			name: 'Profile',
			icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z'
		}
	];

	// Selected menu state
	let currentTab = $state('Dashboard');
	let isMobileSidebarOpen = $state(false);

	// Search & Notification states
	let isNotificationsOpen = $state(false);
	let searchQuery = $state('');

	// Notifications state
	interface NotificationItem {
		id: number;
		text: string;
		time: string;
		unread: boolean;
	}

	let notifications = $state<NotificationItem[]>([]);

	function toggleMobileSidebar() {
		isMobileSidebarOpen = !isMobileSidebarOpen;
	}

	function handleLogout() {
		// Admin Logout -> redirect to home page or admin portal login page
		window.location.href = '/admin-portal';
	}
</script>

<svelte:head>
	<title>Admin Dashboard | iSPARC IIPS DAVV</title>
	<meta
		name="description"
		content="iSPARC Admin Dashboard. Monitor student records, verify certificate submissions, and analyze batch statistics."
	/>
</svelte:head>

<div class="min-h-screen bg-[#F7F6F3] text-slate-800 flex font-sans">
	<!-- ==================== SIDEBAR ==================== -->
	<!-- Desktop Sidebar -->
	<aside
		class="hidden lg:flex flex-col w-64 bg-white border-r border-slate-200 h-screen sticky top-0 shrink-0"
	>
		<!-- Logo area -->
		<div class="h-[72px] flex items-center px-6 gap-3.5 border-b border-slate-100">
			<div
				class="w-9 h-9 bg-[#881B1B] rounded-lg flex items-center justify-center text-white font-bold font-serif text-sm shrink-0"
			>
				A
			</div>
			<div class="flex flex-col">
				<span class="text-xl font-bold tracking-tight text-slate-900 font-serif"
					>i<span class="text-[#881B1B]">SPARC</span></span
				>
				<span class="text-[9px] font-bold text-slate-600 tracking-wider uppercase -mt-1"
					>IIPS DAVV CELL</span
				>
			</div>
		</div>

		<!-- Menu Items -->
		<nav class="flex-grow py-6 px-4 space-y-1 overflow-y-auto">
			{#each menuItems as item}
				<button
					onclick={() => (currentTab = item.name)}
					class="w-full flex items-center gap-3.5 px-4 py-3 rounded-lg text-[13px] font-bold tracking-wide transition-all duration-200 {currentTab ===
					item.name
						? 'bg-[#881B1B]/10 text-[#881B1B] border-l-[3px] border-[#881B1B] rounded-l-none'
						: 'text-slate-700 hover:bg-slate-50 hover:text-slate-900'}"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-5 h-5"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
					</svg>
					{item.name}
				</button>
			{/each}
		</nav>

		<!-- Logout -->
		<div class="p-4 border-t border-slate-100">
			<button
				onclick={handleLogout}
				class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-700 hover:bg-rose-50 hover:text-rose-600 transition-colors"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2"
					class="w-5 h-5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
					/>
				</svg>
				Logout
			</button>
		</div>
	</aside>

	<!-- Mobile Sidebar Overlay Menu (Drawer) -->
	{#if isMobileSidebarOpen}
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			onclick={toggleMobileSidebar}
			transition:fade={{ duration: 150 }}
			class="lg:hidden fixed inset-0 bg-black/40 z-40"
		></div>

		<!-- Drawer Content -->
		<aside
			transition:slide={{ axis: 'x', duration: 250 }}
			class="lg:hidden fixed left-0 top-0 bottom-0 w-64 bg-white z-50 flex flex-col h-full shadow-2xl"
		>
			<div class="h-[72px] flex items-center justify-between px-6 border-b border-slate-100">
				<div class="flex items-center gap-3">
					<div
						class="w-9 h-9 bg-[#881B1B] rounded-lg flex items-center justify-center text-white font-bold font-serif text-sm shrink-0"
					>
						A
					</div>
					<div class="flex flex-col">
						<span class="text-xl font-bold tracking-tight text-slate-900 font-serif"
							>i<span class="text-[#881B1B]">SPARC</span></span
						>
						<span class="text-[9px] font-bold text-slate-600 tracking-wider uppercase -mt-1"
							>IIPS DAVV CELL</span
						>
					</div>
				</div>
				<button
					onclick={toggleMobileSidebar}
					aria-label="Close sidebar"
					class="p-1 rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-6 h-6"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<nav class="flex-grow py-6 px-4 space-y-1 overflow-y-auto">
				{#each menuItems as item}
					<button
						onclick={() => {
							currentTab = item.name;
							toggleMobileSidebar();
						}}
						class="w-full flex items-center gap-3.5 px-4 py-3 rounded-lg text-[13px] font-bold tracking-wide transition-all duration-200 {currentTab ===
						item.name
							? 'bg-[#881B1B]/10 text-[#881B1B] border-l-[3px] border-[#881B1B] rounded-l-none'
							: 'text-slate-700 hover:bg-slate-50 hover:text-slate-900'}"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
							class="w-5 h-5"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
						</svg>
						{item.name}
					</button>
				{/each}
			</nav>

			<div class="p-4 border-t border-slate-100">
				<button
					onclick={() => {
						handleLogout();
						toggleMobileSidebar();
					}}
					class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-700 hover:bg-rose-50 hover:text-rose-600 transition-colors"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
						/>
					</svg>
					Logout
				</button>
			</div>
		</aside>
	{/if}

	<!-- ==================== MAIN WORKSPACE ==================== -->
	<div class="flex-grow flex flex-col min-w-0">
		<!-- Top Navigation Header -->
		<header
			class="bg-white border-b border-slate-200 h-[72px] flex items-center justify-between px-4 sm:px-6 lg:px-8 sticky top-0 z-30"
		>
			<!-- Mobile sidebar toggle -->
			<div class="flex items-center gap-3">
				<button
					onclick={toggleMobileSidebar}
					aria-label="Open sidebar"
					class="lg:hidden p-2 rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-900 focus:outline-none"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-6 h-6"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				</button>
				<div>
					<h1 class="text-lg sm:text-xl font-bold text-slate-900 font-serif leading-tight">
						{currentTab === 'Dashboard'
							? 'Welcome Back, Admin!'
							: currentTab === 'Profile'
								? 'My Profile'
								: currentTab}
					</h1>
					<p
						class="text-[10px] sm:text-xs font-semibold text-slate-400 uppercase tracking-wider mt-0.5"
					>
						{new Date().toLocaleDateString('en-US', {
							weekday: 'long',
							day: 'numeric',
							month: 'long',
							year: 'numeric'
						})}
					</p>
				</div>
			</div>

			<!-- Action items -->
			<div class="flex items-center gap-3.5 relative">
				<!-- Search box -->
				<div class="hidden md:flex items-center relative">
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search students..."
						class="pl-4 pr-9 py-2 bg-slate-50 rounded-lg border border-slate-200 text-xs text-slate-800 focus:outline-none focus:border-slate-350 focus:bg-white w-52 transition-all focus:w-60"
					/>
					<span class="absolute right-3 text-slate-400">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
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

				<!-- Notification button -->
				<button
					onclick={() => {
						isNotificationsOpen = !isNotificationsOpen;
					}}
					class="w-10 h-10 border border-slate-250 bg-white rounded-lg flex items-center justify-center text-slate-500 hover:bg-slate-50 transition-colors relative"
				>
					<span class="sr-only">Notifications</span>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-5 h-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
						/>
					</svg>
					<!-- Active notification indicator (only when unread notifications exist) -->
					{#if notifications.some((n) => n.unread)}
						<span class="absolute top-2 right-2.5 w-2 h-2 bg-[#881B1B] rounded-full"></span>
					{/if}
				</button>

				{#if isNotificationsOpen}
					<div
						transition:slide={{ duration: 150 }}
						class="absolute right-0 top-12 w-80 bg-white border border-slate-200 rounded-xl shadow-lg py-2.5 z-40 mt-1"
					>
						<h3
							class="px-4 py-1.5 text-xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-100"
						>
							Notifications
						</h3>
						<div class="max-h-64 overflow-y-auto">
							{#each notifications as notice}
								<div
									class="px-4 py-3 hover:bg-slate-50 border-b border-slate-50 last:border-b-0 flex gap-2.5"
								>
									{#if notice.unread}
										<span class="w-2.5 h-2.5 bg-[#881B1B] rounded-full mt-1.5 shrink-0"></span>
									{/if}
									<div class="flex-grow">
										<p class="text-xs text-slate-700 font-semibold">{notice.text}</p>
										<span class="text-[10px] text-slate-405 block mt-1">{notice.time}</span>
									</div>
								</div>
							{:else}
								<div class="px-4 py-6 text-center text-xs text-slate-400 font-medium">
									No new notifications
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Profile Pill -->
				<div class="flex items-center gap-3 pl-2 border-l border-slate-200 select-none">
					<div
						class="w-10 h-10 rounded-full bg-[#881B1B] text-white flex items-center justify-center font-bold text-sm border-2 border-white shadow-sm shrink-0 font-serif"
					>
						{admin ? getInitials(admin.name) : 'A'}
					</div>
					<div class="hidden sm:flex flex-col">
						<span class="text-xs font-bold text-slate-900 leading-none">
							{loading ? 'Loading...' : error ? 'Error loading' : admin?.name || 'Administrator'}
						</span>
						<span class="text-[9px] text-slate-400 font-bold tracking-wider block mt-1 uppercase">
							{admin
								? admin.role === 'superadmin'
									? 'Super Administrator'
									: 'Batch Coordinator - ' + admin.assigned_batch
								: 'Coordinator'}
						</span>
					</div>
				</div>
			</div>
		</header>

		<!-- Main Content Body -->
		<main class="flex-grow p-4 sm:p-6 lg:p-8 space-y-6 overflow-y-auto max-w-7xl mx-auto w-full">
			{#if currentTab === 'Dashboard'}
				<AdminDashboardView onTabChange={(tab) => (currentTab = tab)} />
			{:else if currentTab === 'Activity Monitoring'}
				<ActivityMonitoringView />
			{:else if currentTab === 'Student Management'}
				<AdminStudentManagementView />
			{:else if currentTab === 'Certificate Verification'}
				<AdminCertificateVerificationView />
			{:else if currentTab === 'Batch Analytics'}
				<BatchAnalyticsView />
			{:else if currentTab === 'Profile'}
				<AdminProfileView
					{admin}
					{loading}
					{error}
					{stats}
					onEditProfile={() => (currentTab = 'Edit Profile')}
				/>
			{:else if currentTab === 'Edit Profile'}
				<AdminEditProfileView
					{admin}
					onSave={(updatedAdmin, updatedStats) => {
						admin = updatedAdmin;
						if (updatedStats !== undefined) {
							stats = updatedStats;
						}
					}}
					onCancel={() => (currentTab = 'Profile')}
				/>
			{:else}
				<AdminPlaceholderView
					tabName={currentTab}
					onBackToDashboard={() => (currentTab = 'Dashboard')}
				/>
			{/if}
		</main>
	</div>
</div>
