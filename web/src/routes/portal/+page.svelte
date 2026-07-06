<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import { API_BASE_URL } from '$lib/config';
	import ProfileSection from './ProfileSection.svelte';
	import BrowseActivities from './BrowseActivities.svelte';
	import LeaderboardSection from './LeaderboardSection.svelte';
	import CreditsProgress from './CreditsProgress.svelte';
	import ExtracurricularMarksheet from './ExtracurricularMarksheet.svelte';
	import MyEnrollments from './MyEnrollments.svelte';
	import MyCertificates from './MyCertificates.svelte';
	import UploadCertificate from './UploadCertificate.svelte';

	// Sidebar navigation items
	const menuItems = [
		{
			name: 'Dashboard',
			active: true,
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			name: 'Activities',
			active: false,
			icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z'
		},
		{
			name: 'My Enrollments',
			active: false,
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01'
		},
		{
			name: 'Upload Certificate',
			active: false,
			icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12'
		},
		{
			name: 'My Certificates',
			active: false,
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			name: 'Credits & Progress',
			active: false,
			icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		},
		{
			name: 'Leaderboard',
			active: false,
			icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
		},
		{
			name: 'Extracurricular Marksheet',
			active: false,
			icon: 'M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			name: 'Profile',
			active: false,
			icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z'
		}
	];

	// Selected menu item state
	let currentTab = $state('Dashboard');
	let isMobileSidebarOpen = $state(false);

	// Search & Notification states
	let isSearchOpen = $state(false);
	let isNotificationsOpen = $state(false);
	let searchQuery = $state('');

	// Backend API State
	let token = $state('');
	/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
	let student = $state<any>(null);
	let stats = $state({
		activities_participated: 0,
		certificates_uploaded: 0,
		pending_certificates: 0,
		approved_certificates: 0,
		rejected_certificates: 0,
		credits_earned: 0,
		current_rank: 0,
		total_students: 0,
		/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
		recent_activities: [] as any[]
	});
	/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
	let leaderboard = $state<any[]>([]);
	let loading = $state(true);

	// Derived percentage completion and values
	let percentComplete = $derived(Math.min(Math.round((stats.credits_earned / 200) * 100), 100));
	let remainingCredits = $derived(Math.max(200 - stats.credits_earned, 0));
	let approvedPct = $derived(
		stats.certificates_uploaded > 0
			? Math.round((stats.approved_certificates / stats.certificates_uploaded) * 100)
			: 0
	);
	let pendingPct = $derived(
		stats.certificates_uploaded > 0
			? Math.round((stats.pending_certificates / stats.certificates_uploaded) * 100)
			: 0
	);
	let rejectedPct = $derived(
		stats.certificates_uploaded > 0
			? Math.round((stats.rejected_certificates / stats.certificates_uploaded) * 100)
			: 0
	);
	function capitalizeWords(str: string) {
		if (!str) return '';
		return str.toLowerCase().replace(/\b[a-z]/g, (char) => char.toUpperCase());
	}

	let fullName = $derived(student && student.name ? capitalizeWords(student.name) : 'Student');
	let firstName = $derived(
		student && student.name ? capitalizeWords(student.name.trim().split(' ')[0]) : 'Student'
	);
	let totalUploadedPct = $derived(
		stats.activities_participated > 0
			? Math.min(
					Math.round((stats.certificates_uploaded / stats.activities_participated) * 100),
					100
				)
			: 0
	);
	let currentDateString = $derived.by(() => {
		const now = new Date();
		return now.toLocaleDateString('en-GB', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		});
	});

	// Helper to format date
	function formatDate(dateStr: string) {
		if (!dateStr) return '';
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-GB', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		});
	}

	async function loadDashboardData() {
		token = localStorage.getItem('access_token') || '';
		if (!token) {
			window.location.href = '/login';
			return;
		}

		try {
			// Fetch student profile
			const profileRes = await fetch(`${API_BASE_URL}/api/auth/profile`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (!profileRes.ok) {
				localStorage.removeItem('access_token');
				window.location.href = '/login';
				return;
			}

			const profileData = await profileRes.json();
			student = profileData.student;

			// Fetch dashboard statistics
			const statsRes = await fetch(`${API_BASE_URL}/api/student/dashboard/stats`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (statsRes.ok) {
				stats = await statsRes.json();
			}

			// Fetch leaderboard
			const leaderboardRes = await fetch(`${API_BASE_URL}/api/student/leaderboard`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (leaderboardRes.ok) {
				const leaderboardData = await leaderboardRes.json();
				// Map backend list to match frontend layout (limit to top 5)
				/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
				leaderboard = leaderboardData.slice(0, 5).map((item: any, index: number) => {
					const initials = item.name
						.split(' ')
						.map((n: string) => n[0])
						.join('')
						.substring(0, 2)
						.toUpperCase();
					return {
						rank: index + 1,
						initials: initials,
						name: item.name,
						branch: `${item.course_name} . Sem ${item.semester}`,
						points: item.points,
						isSelf: item.is_self
					};
				});
			}
		} catch (err) {
			console.error('Error loading dashboard data:', err);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadDashboardData();
	});

	// Certificate upload simulation state
	let isUploadModalOpen = $state(false);
	let uploadFile = $state<File | null>(null);
	let uploading = $state(false);
	let uploadSuccess = $state(false);

	// Mock Notifications
	const notifications = [
		{
			id: 1,
			text: "Your certificate for 'Robotics Workshop' was audited.",
			time: '2 hours ago',
			unread: true
		},
		{
			id: 2,
			text: 'Debate Championship points approved: +15 credits.',
			time: '1 day ago',
			unread: false
		},
		{
			id: 3,
			text: "New Activity registered: 'Annual Sports Meet 2026'.",
			time: '3 days ago',
			unread: false
		}
	];

	// Derived activities from backend recent_activities
	let activities = $derived(
		/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
		(stats.recent_activities || []).map((act: any) => ({
			name: act.activity_name,
			category: act.activity_category,
			date: formatDate(act.activity_date || act.created_at),
			credits: act.credits,
			status: act.status
		}))
	);

	// Derived filtered activities for simple search
	let filteredActivities = $derived(
		activities.filter(
			(act) =>
				act.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				act.category.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	function toggleMobileSidebar() {
		isMobileSidebarOpen = !isMobileSidebarOpen;
	}

	async function handleLogout() {
		const storedToken = localStorage.getItem('access_token');
		localStorage.removeItem('access_token');
		if (storedToken) {
			try {
				await fetch(`${API_BASE_URL}/api/auth/logout`, {
					method: 'POST',
					headers: {
						Authorization: `Bearer ${storedToken}`
					}
				});
			} catch (e) {
				console.error('Logout request failed:', e);
			}
		}
		window.location.href = '/';
	}

	function triggerAction(actionName: string) {
		if (actionName === 'Upload Certificate') {
			isUploadModalOpen = true;
		} else if (actionName === 'Browse Activities') {
			currentTab = 'Activities';
		} else if (actionName === 'Download Marksheet') {
			currentTab = 'Extracurricular Marksheet';
		} else if (actionName === 'View Progress') {
			currentTab = 'Credits & Progress';
		} else if (actionName === 'View My Certificates') {
			currentTab = 'My Certificates';
		} else {
			alert(`Quick Action Triggered: ${actionName}`);
		}
	}

	function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			uploadFile = target.files[0];
		}
	}

	// Submit modal certificate upload to backend
	async function handleUploadSubmit(event: Event) {
		event.preventDefault();
		if (!uploadFile) return;

		uploading = true;

		const formData = new FormData();
		// Fetch details from inputs
		const formEl = event.target as HTMLFormElement;
		const actNameInput = formEl.querySelector('#certificate-name') as HTMLInputElement;
		const actDateInput = formEl.querySelector('#certificate-date') as HTMLInputElement;
		const orgInput = formEl.querySelector('#certificate-organizer') as HTMLInputElement;
		const levelInput = formEl.querySelector('#certificate-level') as HTMLSelectElement;
		const certNoInput = formEl.querySelector('#certificate-number') as HTMLInputElement;
		const issueDateInput = formEl.querySelector('#certificate-issue-date') as HTMLInputElement;
		const partTypeInput = formEl.querySelector('#certificate-participation') as HTMLSelectElement;
		const descInput = formEl.querySelector('#certificate-desc') as HTMLTextAreaElement;

		formData.append('activity_name', actNameInput?.value || '');
		formData.append('activity_date', actDateInput?.value || '');
		formData.append('organizer_name', orgInput?.value || '');
		formData.append('event_level', levelInput?.value || '');
		formData.append('cert_number', certNoInput?.value || '');
		formData.append('issue_date', issueDateInput?.value || '');
		formData.append('participation_type', partTypeInput?.value || '');
		formData.append('description', descInput?.value || '');
		formData.append('certificate_file', uploadFile);

		// Determine a guessed category based on name
		const lowerName = (actNameInput?.value || '').toLowerCase();
		let activityCategory = 'General / Extracurricular';
		if (lowerName.includes('debate') || lowerName.includes('speech') || lowerName.includes('mun')) {
			activityCategory = 'LITERARY';
		} else if (
			lowerName.includes('dance') ||
			lowerName.includes('drama') ||
			lowerName.includes('singing') ||
			lowerName.includes('music')
		) {
			activityCategory = 'CULTURAL';
		} else if (
			lowerName.includes('hackathon') ||
			lowerName.includes('coding') ||
			lowerName.includes('robotics')
		) {
			activityCategory = 'TECHNICAL';
		} else if (
			lowerName.includes('blood') ||
			lowerName.includes('nss') ||
			lowerName.includes('cleanup')
		) {
			activityCategory = 'SOCIAL SERVICE';
		} else if (
			lowerName.includes('olympiad') ||
			lowerName.includes('quiz') ||
			lowerName.includes('paper')
		) {
			activityCategory = 'RESEARCH';
		} else if (
			lowerName.includes('cricket') ||
			lowerName.includes('football') ||
			lowerName.includes('sports') ||
			lowerName.includes('athletics')
		) {
			activityCategory = 'SPORTS';
		}
		formData.append('activity_category', activityCategory);

		try {
			const res = await fetch(`${API_BASE_URL}/api/student/certificates`, {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${token}`
				},
				body: formData
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'Failed to upload certificate');
			}

			uploadSuccess = true;
			// Refresh statistics
			loadDashboardData();

			setTimeout(() => {
				isUploadModalOpen = false;
				uploadSuccess = false;
				uploadFile = null;
			}, 1500);
		} catch (err) {
			console.error(err);
			alert(err instanceof Error ? err.message : 'Error uploading certificate');
		} finally {
			uploading = false;
		}
	}
</script>

{#if loading}
	<div class="min-h-screen w-full bg-[#F7F6F3] flex items-center justify-center font-sans">
		<div class="flex flex-col items-center gap-4 animate-fade-in">
			<svg class="animate-spin h-8 w-8 text-[#881B1B]" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
				></circle>
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				></path>
			</svg>
			<span class="text-xs font-bold text-slate-500 uppercase tracking-widest animate-pulse"
				>Loading iSPARC Dashboard...</span
			>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-[#F7F6F3] text-slate-800 flex font-sans">
		<!-- ==================== SIDEBAR ==================== -->
		<!-- Desktop Sidebar -->
		<aside
			class="hidden lg:flex flex-col w-64 bg-white border-r border-slate-200 h-screen sticky top-0 shrink-0"
		>
			<!-- Logo area -->
			<div class="h-[72px] flex items-center px-6 gap-3.5 border-b border-slate-100">
				<div
					class="w-8 h-8 bg-[#881B1B] text-white flex items-center justify-center font-bold text-sm rounded"
				>
					S
				</div>
				<div class="flex flex-col">
					<span class="text-xl font-bold tracking-tight text-slate-900 font-serif"
						>i<span class="text-[#881B1B]">SPARC</span></span
					>
					<span class="text-[9px] font-bold text-slate-405 tracking-wider uppercase -mt-1"
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
							: 'text-slate-505 hover:bg-slate-50 hover:text-slate-900'}"
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
					class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-505 hover:bg-rose-50 hover:text-rose-600 transition-colors"
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
							class="w-8 h-8 bg-[#881B1B] text-white flex items-center justify-center font-bold text-sm rounded"
						>
							S
						</div>
						<div class="flex flex-col">
							<span class="text-xl font-bold tracking-tight text-slate-900 font-serif"
								>i<span class="text-[#881B1B]">SPARC</span></span
							>
							<span class="text-[9px] font-bold text-slate-405 tracking-wider uppercase -mt-1"
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
								: 'text-slate-505 hover:bg-slate-50 hover:text-slate-900'}"
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
						class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-505 hover:bg-rose-50 hover:text-rose-600 transition-colors"
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
						class="lg:hidden p-2 rounded-lg text-slate-655 hover:bg-slate-100 hover:text-slate-900 focus:outline-none"
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
							{currentTab === 'Profile'
								? 'My Profile'
								: currentTab === 'Dashboard'
									? `Welcome Back, ${firstName}!`
									: currentTab === 'Activities'
										? 'Browse Activities'
										: currentTab === 'Leaderboard'
											? 'Leader board'
											: currentTab}
						</h1>
						<p
							class="text-[10px] sm:text-xs font-semibold text-slate-405 uppercase tracking-wider mt-0.5"
						>
							{currentDateString}
						</p>
					</div>
				</div>

				<!-- Action items -->
				<div class="flex items-center gap-3.5 relative">
					<!-- Notification button -->
					<button
						onclick={() => {
							isNotificationsOpen = !isNotificationsOpen;
							isSearchOpen = false;
						}}
						class="w-10 h-10 border border-slate-250 bg-white rounded-lg flex items-center justify-center text-slate-655 hover:bg-slate-50 transition-colors relative"
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
						<!-- Active dot indicator -->
						<span class="absolute top-2 right-2.5 w-2 h-2 bg-[#881B1B] rounded-full"></span>
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
											<span class="text-[10px] text-slate-400 block mt-1">{notice.time}</span>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}

					<!-- Search button toggle -->
					<button
						onclick={() => {
							isSearchOpen = !isSearchOpen;
							isNotificationsOpen = false;
						}}
						class="w-10 h-10 border border-slate-250 bg-white rounded-lg flex items-center justify-center text-slate-655 hover:bg-slate-50 transition-colors"
					>
						<span class="sr-only">Toggle search</span>
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
								d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
							/>
						</svg>
					</button>

					{#if isSearchOpen}
						<div
							transition:slide={{ duration: 150 }}
							class="absolute right-0 top-12 w-72 sm:w-80 bg-white border border-slate-200 rounded-xl shadow-lg p-3 z-40 mt-1"
						>
							<div class="relative flex items-center">
								<input
									type="text"
									bind:value={searchQuery}
									placeholder="Search activity or category..."
									class="w-full pl-9 pr-3.5 py-2 bg-slate-50 rounded-lg border border-slate-200 text-xs text-slate-800 focus:outline-none focus:border-slate-350 focus:bg-white"
								/>
								<span class="absolute left-3 text-slate-400">
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
						</div>
					{/if}

					<!-- Profile Pill -->
					<div class="hidden sm:flex items-center gap-3.5 pl-2 border-l border-slate-200">
						<div
							class="w-10 h-10 rounded-full bg-[#881B1B] text-white flex items-center justify-center font-bold text-sm border-2 border-white shadow-sm shrink-0"
						>
							{student
								? student.name
										.split(' ')
										.map((n: string) => n[0])
										.join('')
										.substring(0, 2)
										.toUpperCase()
								: ''}
						</div>
						<div class="flex flex-col">
							<span class="text-xs font-bold text-slate-900 leading-none">{fullName}</span>
							<span
								class="text-[10px] text-slate-405 font-semibold tracking-wide block mt-1 uppercase"
								>{student ? `${student.course_name} · Sem ${student.semester}` : ''}</span
							>
						</div>
					</div>
				</div>
			</header>

			<!-- Main Content Body -->
			<main class="flex-grow p-4 sm:p-6 lg:p-8 space-y-6 overflow-y-auto max-w-7xl mx-auto w-full">
				{#if currentTab === 'Dashboard'}
					<!-- ==================== 1. KPI WIDGETS ROW ==================== -->
					<section
						class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
						aria-label="Key Performance Indicators"
					>
						<!-- Card 1: Activities Participated -->
						<div
							class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs"
						>
							<div class="flex items-center justify-between">
								<span class="text-2xl font-bold font-serif text-slate-900"
									>{stats.activities_participated}</span
								>
								<div class="p-2.5 rounded-lg bg-red-50 text-[#881B1B] border border-red-100">
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
											d="M4 6h16M4 10h16M4 14h16M4 18h16"
										/>
									</svg>
								</div>
							</div>
							<div class="mt-4">
								<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
									Activities Participated
								</h3>
								<p class="text-[10px] font-medium text-slate-400 mt-1 uppercase tracking-wider">
									This academic year
								</p>
							</div>
						</div>

						<!-- Card 2: Certificates Uploaded -->
						<div
							class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs"
						>
							<div class="flex items-center justify-between">
								<span class="text-2xl font-bold font-serif text-slate-900"
									>{stats.certificates_uploaded}</span
								>
								<div class="p-2.5 rounded-lg bg-amber-50 text-[#C89B3C] border border-amber-100">
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
											d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
										/>
									</svg>
								</div>
							</div>
							<div class="mt-4">
								<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
									Certificates Uploaded
								</h3>
								<p class="text-[10px] font-medium text-slate-400 mt-1 uppercase tracking-wider">
									{stats.pending_certificates} pending verification
								</p>
							</div>
						</div>

						<!-- Card 3: Credits Earned -->
						<div
							class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs"
						>
							<div class="flex items-center justify-between">
								<span class="text-2xl font-bold font-serif text-slate-900"
									>{stats.credits_earned}</span
								>
								<div
									class="p-2.5 rounded-lg bg-emerald-50 text-emerald-655 border border-emerald-100"
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
											d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138z"
										/>
									</svg>
								</div>
							</div>
							<div class="mt-4">
								<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
									Credits Earned
								</h3>
								<p class="text-[10px] font-medium text-slate-400 mt-1 uppercase tracking-wider">
									Out of 200 required
								</p>
							</div>
						</div>

						<!-- Card 4: Current Rank -->
						<div
							class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs"
						>
							<div class="flex items-center justify-between">
								<span class="text-2xl font-bold font-serif text-slate-900"
									>#{stats.current_rank}</span
								>
								<div class="p-2.5 rounded-lg bg-amber-50 text-amber-500 border border-amber-100">
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
											d="M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5a2 2 0 10-2 2h2zm0 0h4a2 2 0 012 2v2a2 2 0 01-2 2h-4v-6zm0 0H8a2 2 0 00-2 2v2a2 2 0 002 2h4v-6z"
										/>
									</svg>
								</div>
							</div>
							<div class="mt-4">
								<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
									Current Rank
								</h3>
								<p class="text-[10px] font-medium text-slate-400 mt-1 uppercase tracking-wider">
									Among {stats.total_students} students
								</p>
							</div>
						</div>
					</section>

					<!-- ==================== 2. PROGRESS & STATUS ROW ==================== -->
					<section class="grid grid-cols-1 lg:grid-cols-12 gap-6">
						<!-- Left Panel: Credits Progress -->
						<div
							class="lg:col-span-8 bg-white border border-slate-200 p-6 rounded-xl shadow-xs space-y-6"
						>
							<div class="flex items-center justify-between">
								<div>
									<h2 class="text-base font-bold font-serif text-[#0B1535]">Credits Progress</h2>
									<span
										class="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest block mt-1"
										>ACADEMIC YEAR 2025-26</span
									>
								</div>
								<span
									class="bg-[#881B1B]/10 text-[#881B1B] text-[10px] font-extrabold uppercase px-3 py-1.5 rounded-sm tracking-wider"
								>
									{percentComplete}% Complete
								</span>
							</div>

							<div class="flex flex-col sm:flex-row justify-between items-baseline gap-2">
								<div class="flex items-baseline gap-2">
									<span class="text-4xl font-extrabold text-[#0B1535] leading-none"
										>{stats.credits_earned}</span
									>
									<span class="text-xs font-bold text-slate-405 uppercase tracking-widest"
										>Credits Earned</span
									>
									<span class="text-xl text-slate-300 font-light mx-1">/</span>
									<span class="text-lg font-bold text-slate-500">200</span>
									<span class="text-[10px] font-bold text-slate-405 uppercase tracking-widest"
										>Required</span
									>
								</div>
								<div class="text-right">
									<span class="text-2xl font-extrabold text-slate-800 block"
										>{remainingCredits}</span
									>
									<span
										class="text-[9px] font-bold text-slate-405 uppercase tracking-widest block mt-0.5"
										>Remaining</span
									>
								</div>
							</div>

							<!-- Custom Progress Bar -->
							<div class="space-y-2">
								<div class="h-3 w-full bg-slate-100 rounded-full overflow-hidden relative">
									<div
										class="h-full bg-[#881B1B] rounded-full"
										style="width: {percentComplete}%"
									></div>
								</div>
								<!-- Labels -->
								<div
									class="flex justify-between text-[10px] font-bold text-slate-405 font-sans px-1"
								>
									<span>0</span>
									<span>50</span>
									<span>100</span>
									<span>150</span>
									<span>200</span>
								</div>
							</div>

							<!-- Academic Quote Box -->
							<blockquote class="bg-[#881B1B]/5 border-l-[3.5px] border-[#881B1B] p-4 rounded-r-md">
								<p class="text-xs italic text-slate-700 leading-relaxed font-serif">
									"You're making excellent progress toward your extracurricular goals. Keep it up —
									only {remainingCredits} more credits to go!"
								</p>
							</blockquote>
						</div>

						<!-- Right Panel: Certificate Status Overview -->
						<div
							class="lg:col-span-4 bg-white border border-slate-200 p-6 rounded-xl shadow-xs flex flex-col justify-between"
						>
							<div>
								<h2 class="text-base font-bold font-serif text-[#0B1535]">Certificate Status</h2>
								<span
									class="text-[10px] font-bold text-slate-405 uppercase tracking-widest block mt-1"
									>Verification Overview</span
								>
							</div>

							<div class="space-y-4 my-5 flex-grow flex flex-col justify-center">
								<!-- Row 1: Approved -->
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs font-bold text-slate-700">
										<span
											class="text-[11px] font-extrabold text-emerald-655 uppercase tracking-wider"
											>Approved</span
										>
										<span>{stats.approved_certificates}</span>
									</div>
									<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
										<div
											class="h-full bg-emerald-500 rounded-full"
											style="width: {approvedPct}%"
										></div>
									</div>
								</div>

								<!-- Row 2: Pending Verification -->
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs font-bold text-slate-700">
										<span class="text-[11px] font-extrabold text-amber-600 uppercase tracking-wider"
											>Pending Verification</span
										>
										<span>{stats.pending_certificates}</span>
									</div>
									<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
										<div
											class="h-full bg-amber-500 rounded-full"
											style="width: {pendingPct}%"
										></div>
									</div>
								</div>

								<!-- Row 3: Rejected -->
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs font-bold text-slate-700">
										<span class="text-[11px] font-extrabold text-rose-600 uppercase tracking-wider"
											>Rejected</span
										>
										<span>{stats.rejected_certificates}</span>
									</div>
									<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
										<div
											class="h-full bg-rose-500 rounded-full"
											style="width: {rejectedPct}%"
										></div>
									</div>
								</div>

								<!-- Row 4: Total Uploaded -->
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs font-bold text-slate-700">
										<span class="text-[11px] font-extrabold text-[#0B1535] uppercase tracking-wider"
											>Total Uploaded</span
										>
										<span>{stats.certificates_uploaded}</span>
									</div>
									<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
										<div
											class="h-full bg-[#0B1535] rounded-full"
											style="width: {totalUploadedPct}%"
										></div>
									</div>
								</div>
							</div>
						</div>
					</section>

					<!-- ==================== 3. DATA LISTS & LEADERBOARD ROW ==================== -->
					<section class="grid grid-cols-1 lg:grid-cols-12 gap-6">
						<!-- Left Panel: Recent Activities Table -->
						<div
							class="lg:col-span-8 bg-white border border-slate-200 p-5 rounded-xl shadow-xs flex flex-col justify-between min-h-[400px]"
						>
							<div>
								<div class="flex items-center justify-between pb-4 border-b border-slate-100">
									<h2 class="text-base font-bold font-serif text-[#0B1535]">Recent Activities</h2>
									<button
										onclick={() => alert('Redirecting to full activities view...')}
										class="text-[11px] font-bold text-[#881B1B] hover:text-[#731717] uppercase tracking-wider flex items-center gap-1.5 transition-colors"
									>
										View All
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											stroke-width="2.5"
											class="w-3 h-3"
										>
											<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
										</svg>
									</button>
								</div>

								<!-- Search Filter Bar -->
								<div class="mt-4 flex flex-col sm:flex-row items-center gap-3">
									<div class="relative w-full flex items-center">
										<input
											type="text"
											bind:value={searchQuery}
											placeholder="Filter by activity name or category..."
											class="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-350 focus:bg-white"
										/>
										<span class="absolute left-3 text-slate-405">
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
								</div>

								<!-- Activities Table -->
								<div class="overflow-x-auto mt-4">
									<table class="w-full text-left border-collapse text-xs">
										<thead>
											<tr
												class="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest border-b border-slate-100 bg-slate-50/50"
											>
												<th class="py-3 px-4">Activity</th>
												<th class="py-3 px-4">Category</th>
												<th class="py-3 px-4">Date</th>
												<th class="py-3 px-4 text-center">Credits</th>
												<th class="py-3 px-4 text-right">Status</th>
											</tr>
										</thead>
										<tbody class="divide-y divide-slate-100">
											{#if filteredActivities.length > 0}
												{#each filteredActivities as activity}
													<tr class="hover:bg-slate-50/80 transition-colors">
														<td
															class="py-3 px-4 font-bold text-[#0B1535] leading-relaxed max-w-[180px] sm:max-w-none truncate sm:whitespace-normal"
														>
															{activity.name}
														</td>
														<td class="py-3 px-4">
															<span
																class="inline-flex px-2 py-1 bg-slate-100 border border-slate-200/50 text-slate-655 rounded-sm text-[10px] font-bold"
															>
																{activity.category}
															</span>
														</td>
														<td class="py-3 px-4 text-[#6B7280] font-medium whitespace-nowrap"
															>{activity.date}</td
														>
														<td class="py-3 px-4 text-center font-bold text-[#0B1535]"
															>{activity.credits}</td
														>
														<td class="py-3 px-4 text-right whitespace-nowrap">
															<span
																class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-sm text-[10px] font-bold tracking-wide uppercase border {activity.status ===
																'Approved'
																	? 'bg-emerald-50 text-emerald-700 border-emerald-200'
																	: activity.status === 'Pending'
																		? 'bg-amber-50 text-amber-700 border-amber-200'
																		: 'bg-rose-50 text-rose-700 border-rose-200'}"
															>
																<span
																	class="w-1.5 h-1.5 rounded-full {activity.status === 'Approved'
																		? 'bg-emerald-500'
																		: activity.status === 'Pending'
																			? 'bg-amber-500'
																			: 'bg-rose-500'}"
																></span>
																{activity.status}
															</span>
														</td>
													</tr>
												{/each}
											{:else}
												<tr>
													<td
														colspan="5"
														class="py-12 text-center text-slate-405 font-semibold font-sans"
													>
														No activities found matching "{searchQuery}"
													</td>
												</tr>
											{/if}
										</tbody>
									</table>
								</div>
							</div>
						</div>

						<!-- Right Panel: Leaderboard widget -->
						<div
							class="lg:col-span-4 bg-white border border-slate-200 p-5 rounded-xl shadow-xs flex flex-col justify-between min-h-[400px]"
						>
							<div>
								<div class="flex items-center justify-between pb-4 border-b border-slate-100">
									<h2 class="text-base font-bold font-serif text-[#0B1535]">Leaderboard</h2>
									<span
										class="inline-flex px-2 py-1 bg-amber-50 text-amber-700 border border-amber-200/40 rounded-sm text-[9px] font-bold uppercase tracking-wider"
									>
										This Year
									</span>
								</div>

								<!-- Leaderboard List -->
								<div class="space-y-2.5 mt-4">
									{#each leaderboard as student}
										<div
											class="flex items-center justify-between py-2 px-3.5 border rounded-lg transition-all duration-200 {student.isSelf
												? 'border-[#881B1B] bg-[#881B1B]/5 ring-1 ring-[#881B1B]/20 font-bold'
												: 'border-slate-150 bg-white hover:border-slate-300'}"
										>
											<div class="flex items-center gap-3">
												<!-- Rank Indicator -->
												<span
													class="w-5.5 h-5.5 rounded-full flex items-center justify-center text-[10px] font-extrabold {student.rank ===
													1
														? 'bg-amber-100 text-amber-800'
														: student.rank === 2
															? 'bg-slate-100 text-slate-700'
															: student.rank === 3
																? 'bg-orange-50 text-orange-800'
																: 'text-slate-405'}"
												>
													{student.rank}
												</span>
												<!-- Initial Avatar -->
												<div
													class="w-8 h-8 rounded-full bg-slate-100 border border-slate-200 text-[#0B1535] text-[11px] font-bold flex items-center justify-center shrink-0"
												>
													{student.initials}
												</div>
												<div>
													<span class="text-xs font-bold text-slate-900 block leading-tight"
														>{student.name}</span
													>
													<span class="text-[9px] text-[#6B7280] font-medium block mt-0.5"
														>{student.branch}</span
													>
												</div>
											</div>
											<span class="text-xs font-extrabold text-[#0B1535]"
												>{student.points}
												<span
													class="text-[9px] text-slate-405 font-semibold uppercase tracking-wider"
													>pts</span
												></span
											>
										</div>
									{/each}
								</div>
							</div>

							<button
								onclick={() => (currentTab = 'Leaderboard')}
								class="w-full text-center py-3 border border-slate-200 bg-slate-50 hover:bg-slate-100 text-[11px] font-extrabold text-[#0B1535] uppercase tracking-wider rounded-lg transition-colors duration-200 mt-4 flex items-center justify-center gap-2"
							>
								View Full Leaderboard
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									stroke-width="2.5"
									class="w-3.5 h-3.5"
								>
									<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
								</svg>
							</button>
						</div>
					</section>

					<!-- ==================== 4. QUICK ACTIONS SECTION ==================== -->
					<section class="space-y-3.5 border-t border-slate-200 pt-6">
						<h2 class="text-base font-bold font-serif text-[#0B1535]">Quick Actions</h2>

						<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
							<!-- Action 1: Upload Certificate -->
							<button
								onclick={() => triggerAction('Upload Certificate')}
								class="flex flex-col items-center justify-center text-center p-5 bg-[#881B1B] hover:bg-[#731717] text-white rounded-xl shadow-xs transition-all duration-200 group focus:outline-none focus:ring-2 focus:ring-[#881B1B] focus:ring-offset-2"
							>
								<div
									class="w-10 h-10 bg-white/10 rounded-full flex items-center justify-center mb-3 group-hover:scale-105 transition-transform"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										stroke-width="2.2"
										class="w-5 h-5 text-white"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
										/>
									</svg>
								</div>
								<span class="text-xs font-bold tracking-wide">Upload Certificate</span>
							</button>

							<!-- Action 2: View Certificates -->
							<button
								onclick={() => triggerAction('View My Certificates')}
								class="flex flex-col items-center justify-center text-center p-5 bg-white border border-slate-200 hover:border-slate-350 hover:bg-slate-50 text-slate-655 rounded-xl shadow-xs transition-all duration-200 group focus:outline-none focus:ring-2 focus:ring-[#881B1B] focus:ring-offset-2"
							>
								<div
									class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center mb-3 group-hover:scale-105 transition-transform"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										stroke-width="2.2"
										class="w-5 h-5 text-slate-600"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
										/>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
										/>
									</svg>
								</div>
								<span class="text-xs font-bold tracking-wide text-slate-800"
									>View My Certificates</span
								>
							</button>

							<!-- Action 3: Browse Activities -->
							<button
								onclick={() => triggerAction('Browse Activities')}
								class="flex flex-col items-center justify-center text-center p-5 bg-white border border-slate-200 hover:border-slate-350 hover:bg-slate-50 text-slate-655 rounded-xl shadow-xs transition-all duration-200 group focus:outline-none focus:ring-2 focus:ring-[#881B1B] focus:ring-offset-2"
							>
								<div
									class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center mb-3 group-hover:scale-105 transition-transform"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										stroke-width="2.2"
										class="w-5 h-5 text-slate-600"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
										/>
									</svg>
								</div>
								<span class="text-xs font-bold tracking-wide text-slate-800">Browse Activities</span
								>
							</button>

							<!-- Action 4: View Progress -->
							<button
								onclick={() => triggerAction('View Progress')}
								class="flex flex-col items-center justify-center text-center p-5 bg-white border border-slate-200 hover:border-slate-350 hover:bg-slate-50 text-slate-655 rounded-xl shadow-xs transition-all duration-200 group focus:outline-none focus:ring-2 focus:ring-[#881B1B] focus:ring-offset-2"
							>
								<div
									class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center mb-3 group-hover:scale-105 transition-transform"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										stroke-width="2.2"
										class="w-5 h-5 text-slate-600"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
										/>
									</svg>
								</div>
								<span class="text-xs font-bold tracking-wide text-slate-800">View Progress</span>
							</button>

							<!-- Action 5: Download Marksheet -->
							<button
								onclick={() => triggerAction('Download Marksheet')}
								class="flex flex-col items-center justify-center text-center p-5 bg-white border border-slate-200 hover:border-slate-350 hover:bg-slate-50 text-slate-655 rounded-xl shadow-xs transition-all duration-200 group focus:outline-none focus:ring-2 focus:ring-[#881B1B] focus:ring-offset-2"
							>
								<div
									class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center mb-3 group-hover:scale-105 transition-transform"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										stroke-width="2.2"
										class="w-5 h-5 text-slate-600"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
										/>
									</svg>
								</div>
								<span class="text-xs font-bold tracking-wide text-slate-800"
									>Download Marksheet</span
								>
							</button>
						</div>
					</section>
				{:else if currentTab === 'Profile'}
					<ProfileSection />
				{:else if currentTab === 'Activities'}
					<BrowseActivities />
				{:else if currentTab === 'Leaderboard'}
					<LeaderboardSection />
				{:else if currentTab === 'Credits & Progress'}
					<CreditsProgress />
				{:else if currentTab === 'Extracurricular Marksheet'}
					<ExtracurricularMarksheet />
				{:else if currentTab === 'My Enrollments'}
					<MyEnrollments onUploadCertificate={() => (isUploadModalOpen = true)} />
				{:else if currentTab === 'My Certificates'}
					<MyCertificates
						onTabChange={(tab) => (currentTab = tab)}
						onSelectForReupload={() => {}}
					/>
				{:else if currentTab === 'Upload Certificate'}
					<UploadCertificate onBackToDashboard={() => (currentTab = 'Dashboard')} />
				{:else}
					<!-- Placeholder for under construction pages -->
					<div
						transition:fade={{ duration: 150 }}
						class="bg-white border border-slate-200 rounded-xl p-8 sm:p-12 text-center max-w-2xl mx-auto my-12 shadow-xs space-y-6 animate-fade-in"
					>
						<div
							class="w-16 h-16 bg-[#881B1B]/10 text-[#881B1B] flex items-center justify-center rounded-full mx-auto"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								stroke-width="2"
								class="w-8 h-8 animate-bounce"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
								/>
							</svg>
						</div>

						<div class="space-y-2">
							<h2 class="text-xl font-bold font-serif text-[#0B1535]">
								{currentTab} Section Under Construction
							</h2>
							<p class="text-xs text-slate-500 max-w-md mx-auto leading-relaxed font-sans">
								The features for the <strong>{currentTab}</strong> module are scheduled for development
								in the next release phase. Please return to the primary dashboard.
							</p>
						</div>

						<div class="pt-2">
							<button
								onclick={() => (currentTab = 'Dashboard')}
								class="inline-flex items-center justify-center bg-[#881B1B] hover:bg-[#731717] text-white font-bold text-xs tracking-wider uppercase px-6 py-3.5 border border-[#881B1B] transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-[#881B1B] rounded-lg cursor-pointer"
							>
								Return to Dashboard
							</button>
						</div>
					</div>
				{/if}
			</main>
		</div>
	</div>

	<!-- ==================== 5. MODALS & OVERLAYS ==================== -->
	{#if isUploadModalOpen}
		<!-- Backdrop -->
		<div
			transition:fade={{ duration: 150 }}
			class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
		>
			<!-- Modal Card -->
			<div
				transition:slide={{ duration: 250 }}
				class="bg-white border border-slate-250 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden font-sans"
				role="dialog"
				aria-modal="true"
				aria-labelledby="upload-modal-title"
			>
				<div class="h-1 bg-[#881B1B]"></div>

				<!-- Header -->
				<div class="p-6 border-b border-slate-100 flex items-center justify-between">
					<div>
						<h3 id="upload-modal-title" class="text-lg font-bold font-serif text-[#0B1535]">
							Upload Activity Certificate
						</h3>
						<p class="text-[10px] font-semibold text-[#6B7280] uppercase tracking-widest mt-0.5">
							student score registry portal
						</p>
					</div>
					<button
						onclick={() => (isUploadModalOpen = false)}
						aria-label="Close modal"
						class="p-1 rounded-lg text-slate-400 hover:bg-slate-50 hover:text-slate-600 transition-colors"
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

				<!-- Form -->
				<form onsubmit={handleUploadSubmit} class="p-6 space-y-4">
					{#if uploadSuccess}
						<div transition:fade={{ duration: 250 }} class="text-center py-8 space-y-4">
							<div
								class="w-14 h-14 bg-emerald-100 text-emerald-800 flex items-center justify-center rounded-full mx-auto border-2 border-emerald-250 animate-bounce"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									stroke-width="3"
									class="w-7 h-7"
								>
									<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
								</svg>
							</div>
							<div>
								<h4 class="text-base font-bold text-slate-900">Certificate Uploaded!</h4>
								<p class="text-xs text-slate-500 mt-1">
									Undergoing audit by iSPARC cell coordinator.
								</p>
							</div>
						</div>
					{:else}
						<!-- File selection -->
						<div class="flex flex-col gap-1.5">
							<label
								for="certificate-file"
								class="text-[11px] font-bold text-slate-700 tracking-wider"
								>SELECT FILE (PDF, JPG or PNG) <span class="text-[#881B1B]">*</span></label
							>
							<div
								class="border-2 border-dashed border-slate-200 hover:border-slate-350 rounded-xl p-6 text-center transition-colors relative flex flex-col items-center justify-center bg-slate-50/50"
							>
								<input
									id="certificate-file"
									type="file"
									accept=".pdf,.jpg,.jpeg,.png"
									onchange={handleFileChange}
									class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
									required
								/>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									stroke-width="2"
									class="w-8 h-8 text-slate-400 mb-2"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
									/>
								</svg>
								{#if uploadFile}
									<span class="text-xs font-bold text-[#881B1B]">{uploadFile.name}</span>
									<span class="text-[10px] text-[#6B7280] block mt-1"
										>{(uploadFile.size / 1024 / 1024).toFixed(2)} MB</span
									>
								{:else}
									<span class="text-xs font-semibold text-slate-600"
										>Drag & drop your file or click to browse</span
									>
									<span class="text-[10px] text-slate-400 block mt-1">Maximum upload size: 5MB</span
									>
								{/if}
							</div>
						</div>

						<!-- Description -->
						<div class="flex flex-col gap-1.5">
							<label
								for="certificate-desc"
								class="text-[11px] font-bold text-slate-700 tracking-wider"
								>ACTIVITY DETAILS & DESCRIPTION <span class="text-[#881B1B]">*</span></label
							>
							<textarea
								id="certificate-desc"
								rows="3"
								placeholder="Provide context about your involvement, ranks won, or dates of service..."
								class="w-full px-3.5 py-2.5 bg-white rounded-lg border border-slate-250 text-[13px] text-slate-800 placeholder:text-slate-400 focus:outline-none focus:border-[#881B1B] focus:ring-2 focus:ring-[#881B1B]/10 transition-all duration-200"
								required></textarea>
						</div>

						<!-- Buttons -->
						<div class="flex items-center gap-3 pt-3">
							<button
								type="button"
								onclick={() => (isUploadModalOpen = false)}
								class="flex-1 py-3 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs tracking-wider uppercase rounded-xl transition duration-200"
							>
								Cancel
							</button>
							<button
								type="submit"
								disabled={uploading || !uploadFile}
								class="flex-grow py-3 bg-[#881B1B] hover:bg-[#731717] disabled:bg-slate-200 disabled:text-slate-400 text-white font-bold text-xs tracking-wider uppercase rounded-xl transition duration-200 flex items-center justify-center gap-2"
							>
								{#if uploading}
									<svg class="animate-spin h-4.5 w-4.5 text-white" fill="none" viewBox="0 0 24 24">
										<circle
											class="opacity-25"
											cx="12"
											cy="12"
											r="10"
											stroke="currentColor"
											stroke-width="4"
										></circle>
										<path
											class="opacity-75"
											fill="currentColor"
											d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
										></path>
									</svg>
									Uploading...
								{:else}
									Submit Certificate
								{/if}
							</button>
						</div>
					{/if}
				</form>
			</div>
		</div>
	{/if}
{/if}
