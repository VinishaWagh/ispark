<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { API_BASE_URL } from '$lib/config';
	import ActivityManagementView from './ActivityManagementView.svelte';
	import TrackManagementView from './TrackManagementView.svelte';
	import AnnouncementManagementView from '../AnnouncementManagementView.svelte';

	// ── Platform data (loaded from the API) ────────────────────────────────────
	interface PlatformUser {
		name: string;
		id: string;
		role: string;
		displayRole: string;
		dept: string;
		status: string;
		avatarBg: string;
		email?: string;
		semester?: number;
	}

	interface PlatformStats {
		total_students: number;
		total_admins: number;
		total_users: number;
		total_activities: number;
		total_certificates: number;
		pending_certificates: number;
		active_tracks: number;
	}

	const avatarPalette = [
		'bg-[#802D2D] text-white',
		'bg-[#2A4D69] text-white',
		'bg-[#7D4427] text-white',
		'bg-[#244E3F] text-white',
		'bg-[#3C4A8E] text-white',
		'bg-[#6B5E33] text-white'
	];

	let stats = $state<PlatformStats>({
		total_students: 0,
		total_admins: 0,
		total_users: 0,
		total_activities: 0,
		total_certificates: 0,
		pending_certificates: 0,
		active_tracks: 0
	});
	let loadError = $state('');

	function authHeaders(): Record<string, string> {
		return { Authorization: `Bearer ${localStorage.getItem('superadmin_token')}` };
	}

	async function loadPlatformData() {
		try {
			const [statsRes, usersRes] = await Promise.all([
				fetch(`${API_BASE_URL}/api/admin/platform/stats`, { headers: authHeaders() }),
				fetch(`${API_BASE_URL}/api/admin/platform/users`, { headers: authHeaders() })
			]);

			if (statsRes.status === 401 || usersRes.status === 401) {
				localStorage.removeItem('superadmin_token');
				await goto('/super-admin-portal');
				return;
			}

			if (!statsRes.ok || !usersRes.ok) {
				throw new Error('Failed to load platform data');
			}

			stats = await statsRes.json();

			const { users } = await usersRes.json();
			userRegistry = users.map(
				(
					user: { name: string; id: string; role: string; dept: string; status: string },
					i: number
				) => ({
					...user,
					displayRole: user.role,
					avatarBg: avatarPalette[i % avatarPalette.length]
				})
			);
			loadError = '';
		} catch {
			loadError = 'Could not load platform data. Please try again.';
		}
	}

	onMount(async () => {
		// The dashboard is reachable by URL, so it must not render for a visitor
		// who never signed in.
		if (!localStorage.getItem('superadmin_token')) {
			await goto('/super-admin-portal');
			return;
		}

		await Promise.all([loadPlatformData(), loadSettings(), loadReportsData()]);
	});

	// Sidebar menu items list for Super Admin Portal
	const menuItems = [
		{
			name: 'Dashboard',
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			name: 'User Management',
			icon: 'M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75'
		},
		{
			name: 'Activity Management',
			icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		},
		{
			name: 'Track Management',
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			name: 'Announcement Management',
			icon: 'M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z'
		},
		{
			name: 'Reports Center',
			icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z'
		},
		{
			name: 'System Settings',
			icon: 'M12 15.5A3.5 3.5 0 1012 8a3.5 3.5 0 000 7.5zM19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 11-4 0v-.09a1.65 1.65 0 00-1.08-1.5 1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06A1.65 1.65 0 004.6 15a1.65 1.65 0 00-1.51-1H3a2 2 0 110-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 112.83-2.83l.06.06A1.65 1.65 0 009 4.6h.01A1.65 1.65 0 0010 3.09V3a2 2 0 114 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 112.83 2.83l-.06.06A1.65 1.65 0 0019.4 9v.01A1.65 1.65 0 0020.91 10H21a2 2 0 110 4h-.09A1.65 1.65 0 0019.4 15z'
		}
	];

	// Selected menu state
	let currentTab = $state('Dashboard');
	let isMobileSidebarOpen = $state(false);

	// Search & Notification states
	let isNotificationsOpen = $state(false);
	let searchQuery = $state('');

	// Mock Notifications for Super Admin
	const notifications = [
		{
			id: 1,
			text: 'System health warning: Background DB sync delayed by 3m.',
			time: '15 mins ago',
			unread: true
		},
		{
			id: 2,
			text: 'Security Alert: Failed SSH attempt from 192.168.1.104.',
			time: '1 hour ago',
			unread: true
		},
		{
			id: 3,
			text: 'Daily audit backup successfully uploaded to S3.',
			time: '6 hours ago',
			unread: false
		}
	];

	// Users come from the API. recentUsers is just the newest slice of the same
	// registry, so the two can never drift apart.
	let userRegistry = $state<PlatformUser[]>([]);
	const recentUsers = $derived(userRegistry.slice(0, 5));

	let userFilter = $state('All');
	let userSearch = $state('');

	let filteredUsers = $derived(
		userRegistry.filter((user) => {
			const matchesFilter =
				userFilter === 'All' ||
				(userFilter === 'Student' && user.role === 'Student') ||
				(userFilter === 'Admin' &&
					(user.displayRole === 'Admin' || user.displayRole === 'Super Admin'));

			const matchesSearch =
				user.name.toLowerCase().includes(userSearch.toLowerCase()) ||
				user.id.toLowerCase().includes(userSearch.toLowerCase()) ||
				user.dept.toLowerCase().includes(userSearch.toLowerCase());

			return matchesFilter && matchesSearch;
		})
	);

	// Counted from the real registry. These used to add a fixed number to each
	// count, which made every figure on this page fictional.
	let totalStudentsCount = $derived(userRegistry.filter((u) => u.role === 'Student').length);
	let totalAdminsCount = $derived(
		userRegistry.filter((u) => u.displayRole === 'Admin' || u.displayRole === 'Super Admin').length
	);
	let activeUsersCount = $derived(userRegistry.filter((u) => u.status === 'Active').length);
	let pendingUsersCount = $derived(userRegistry.filter((u) => u.status === 'Pending').length);
	let inactiveUsersCount = $derived(userRegistry.filter((u) => u.status === 'Inactive').length);
	let activeRate = $derived(
		userRegistry.length > 0 ? ((activeUsersCount / userRegistry.length) * 100).toFixed(1) : '0.0'
	);

	// Mock Recent System Activities (Step 5)
	let recentLogs = $state([
		{
			activity: 'New student account created',
			type: 'User Added',
			performedBy: 'Super Admin',
			date: 'Jun 27, 2026',
			status: 'Completed'
		},
		{
			activity: 'Activity "Blood Donation Camp" published',
			type: 'Activity Created',
			performedBy: 'Super Admin',
			date: 'Jun 25, 2026',
			status: 'Completed'
		},
		{
			activity: 'Track "Interdisciplinary" updated',
			type: 'Track Updated',
			performedBy: 'Super Admin',
			date: 'Jun 24, 2026',
			status: 'Completed'
		},
		{
			activity: 'Platform-wide announcement published',
			type: 'Announcement',
			performedBy: 'Super Admin',
			date: 'Jun 22, 2026',
			status: 'Completed'
		},
		{
			activity: 'Monthly activity report generated',
			type: 'Report',
			performedBy: 'Super Admin',
			date: 'Jun 20, 2026',
			status: 'Completed'
		},
		{
			activity: 'Admin assigned to Data Science track',
			type: 'Track Updated',
			performedBy: 'Super Admin',
			date: 'Jun 18, 2026',
			status: 'Completed'
		}
	]);

	// Action modals states (Step 6)
	let isCreateUserModalOpen = $state(false);
	let isCreateActivityModalOpen = $state(false);
	let isManageActivityModalOpen = $state(false);
	let isCreateTrackModalOpen = $state(false);
	let isGenerateReportModalOpen = $state(false);

	// Action form states
	let newUserName = $state('');
	let newUserRole = $state('Student');
	let newUserDept = $state('');
	let newUserId = $state('');
	let newUserEmail = $state('');
	let newUserSemester = $state(1);
	let createUserError = $state('');
	let creatingUser = $state(false);

	// View / Edit user modal states
	let isViewUserModalOpen = $state(false);
	let viewingUser = $state<PlatformUser | null>(null);

	let isEditUserModalOpen = $state(false);
	let editingUser = $state<PlatformUser | null>(null);

	let editUserName = $state('');
	let editUserEmail = $state('');
	let editUserDept = $state('');
	let editUserSemester = $state(1);
	let editUserStatus = $state('Active');
	let editUserError = $state('');
	let editingUserSubmit = $state(false);

	function openViewUser(user: PlatformUser) {
		viewingUser = user;
		isViewUserModalOpen = true;
	}

	function openEditUser(user: PlatformUser) {
		editingUser = user;
		editUserName = user.name;
		editUserEmail = user.email || '';
		editUserDept = user.dept;
		editUserSemester = user.semester || 1;
		editUserStatus = user.status || 'Active';
		editUserError = '';
		isEditUserModalOpen = true;
	}

	async function handleSaveEditUser(e: Event) {
		e.preventDefault();
		if (!editingUser || editingUserSubmit) return;

		editUserError = '';
		editingUserSubmit = true;

		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/users/${editingUser.id}`, {
				method: 'PUT',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: editUserName,
					email: editUserEmail,
					dept: editUserDept,
					semester: Number(editUserSemester),
					status: editUserStatus
				})
			});

			let data = { error: '' };
			try {
				data = await res.json();
			} catch (err) {
				console.warn(err);
			}
			if (!res.ok) {
				throw new Error(data.error || 'Failed to update user');
			}

			triggerToast(`User "${editUserName}" updated successfully!`);
			await loadPlatformData();
			isEditUserModalOpen = false;
		} catch (err) {
			editUserError = err instanceof Error ? err.message : 'Failed to update user';
		} finally {
			editingUserSubmit = false;
		}
	}
	let newActivityName = $state('');
	let newActivityCredits = $state(5);
	let newTrackName = $state('');

	// ── System Settings (Step 7) ────────────────────────────────────────────────
	type SettingStatus = 'Active' | 'Enabled' | 'Disabled';
	interface PlatformSetting {
		key: string;
		name: string;
		description: string;
		value: string;
		status: SettingStatus;
	}

	const settingsTabs = [
		'Academic Year',
		'Credit Policy',
		'Activity Rules',
		'Notifications',
		'Platform',
		'Security'
	];

	let activeSettingsTab = $state('Academic Year');

	// Settings are loaded from the API on mount; see loadSettings().
	let settingsSections = $state<Record<string, PlatformSetting[]>>({});

	async function loadSettings() {
		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/settings`, {
				headers: authHeaders()
			});

			if (res.status === 401) {
				localStorage.removeItem('superadmin_token');
				await goto('/super-admin-portal');
				return;
			}

			if (!res.ok) {
				throw new Error('Failed to load settings');
			}

			const { settings } = await res.json();
			settingsSections = settings ?? {};
		} catch {
			loadError = 'Could not load platform data. Please try again.';
		}
	}

	let currentSettings = $derived(settingsSections[activeSettingsTab] ?? []);

	// Edit setting modal state
	let isEditSettingOpen = $state(false);
	let editSettingIndex = $state(-1);
	let editSettingName = $state('');
	let editSettingValue = $state('');

	function openEditSetting(index: number) {
		const s = settingsSections[activeSettingsTab][index];
		editSettingIndex = index;
		editSettingName = s.name;
		editSettingValue = s.value;
		isEditSettingOpen = true;
	}

	async function handleSaveSetting(e: Event) {
		e.preventDefault();
		if (editSettingIndex < 0 || !editSettingValue.trim()) return;

		const setting = settingsSections[activeSettingsTab][editSettingIndex];
		const newValue = editSettingValue.trim();

		try {
			const res = await fetch(
				`${API_BASE_URL}/api/admin/platform/settings/${encodeURIComponent(setting.key)}`,
				{
					method: 'PUT',
					headers: { ...authHeaders(), 'Content-Type': 'application/json' },
					body: JSON.stringify({ value: newValue })
				}
			);

			if (res.status === 401) {
				localStorage.removeItem('superadmin_token');
				await goto('/super-admin-portal');
				return;
			}

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to update setting');
				return;
			}

			settingsSections[activeSettingsTab][editSettingIndex].value = newValue;
			triggerToast(`"${editSettingName}" updated successfully!`);
			isEditSettingOpen = false;
			editSettingIndex = -1;
		} catch {
			triggerToast('Failed to update setting');
		}
	}

	async function handleSaveAllSettings() {
		const tab = activeSettingsTab;
		const payload = {
			settings: (settingsSections[tab] ?? []).map((s) => ({
				key: s.key,
				value: s.value,
				status: s.status
			}))
		};

		if (payload.settings.length === 0) return;

		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/settings`, {
				method: 'PUT',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});

			if (res.status === 401) {
				localStorage.removeItem('superadmin_token');
				await goto('/super-admin-portal');
				return;
			}

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to save settings');
				return;
			}

			const { settings } = await res.json();
			settingsSections = settings ?? settingsSections;
			triggerToast('All system settings saved successfully!');
		} catch {
			triggerToast('Failed to save settings');
		}
	}

	// There is no defaults endpoint, so "reset" reloads the persisted values from
	// the server, discarding any unsaved in-memory edits.
	async function handleResetSettings() {
		await loadSettings();
		triggerToast(`"${activeSettingsTab}" settings reloaded from server.`);
	}

	function settingStatusClass(status: SettingStatus): string {
		return status === 'Disabled'
			? 'bg-slate-100 text-slate-500 border-slate-200'
			: 'bg-emerald-50 text-emerald-700 border-emerald-100';
	}

	function settingStatusDot(status: SettingStatus): string {
		return status === 'Disabled' ? 'bg-slate-400' : 'bg-emerald-500';
	}

	// Toast states
	interface Toast {
		id: number;
		message: string;
	}
	let toasts = $state<Toast[]>([]);
	let toastCounter = 0;

	function triggerToast(message: string, duration = 3000) {
		const id = toastCounter++;
		toasts = [...toasts, { id, message }];
		if (duration > 0) {
			setTimeout(() => {
				toasts = toasts.filter((t) => t.id !== id);
			}, duration);
		}
	}

	function dismissToast(id: number) {
		toasts = toasts.filter((t) => t.id !== id);
	}

	async function handleCreateUser(e: Event) {
		e.preventDefault();
		if (creatingUser) return;

		createUserError = '';
		creatingUser = true;

		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/users`, {
				method: 'POST',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: newUserName,
					role: newUserRole,
					id: newUserId,
					email: newUserEmail,
					dept: newUserDept,
					semester: newUserRole === 'Student' ? Number(newUserSemester) : 0
				})
			});

			let data = { error: '', temporary_password: '' };
			try {
				data = await res.json();
			} catch (err) {
				console.warn(err);
			}

			if (!res.ok) {
				throw new Error(data.error || 'Failed to create user');
			}

			// Reload from the server rather than guessing what was stored, so the
			// registry and the stat cards always agree with the database.
			await loadPlatformData();

			recentLogs = [
				{
					activity: `New ${newUserRole.toLowerCase()} account created for ${newUserName}`,
					type: 'User Added',
					performedBy: 'Super Admin',
					date: new Date().toLocaleDateString('en-GB', {
						day: '2-digit',
						month: 'short',
						year: 'numeric'
					}),
					status: 'Completed'
				},
				...recentLogs
			];

			// Shown once: the account has no other way to get its first password.
			if (newUserRole === 'Student') {
				triggerToast(
					`"${newUserName}" created. Temporary password: ${data.temporary_password} (Student must verify email OTP on first login)`,
					0
				);
			} else {
				triggerToast(`"${newUserName}" created. Temporary password: ${data.temporary_password}`, 0);
			}

			newUserName = '';
			newUserDept = '';
			newUserId = '';
			newUserEmail = '';
			newUserSemester = 1;
			isCreateUserModalOpen = false;
		} catch (err) {
			createUserError = err instanceof Error ? err.message : 'Failed to create user';
		} finally {
			creatingUser = false;
		}
	}

	async function handleDeleteUser(user: PlatformUser) {
		if (!confirm(`Are you sure you want to delete ${user.name}?`)) return;

		try {
			const res = await fetch(`${API_BASE_URL}/api/admin/platform/users/${user.id}`, {
				method: 'DELETE',
				headers: authHeaders()
			});

			let data = { error: '' };
			try {
				data = await res.json();
			} catch (err) {
				console.warn(err);
			}

			if (!res.ok) {
				throw new Error(data.error || 'Failed to delete user');
			}

			await loadPlatformData();
			triggerToast(`User "${user.name}" removed successfully.`);
		} catch (err) {
			triggerToast(err instanceof Error ? err.message : 'Failed to delete user');
		}
	}

	function handleCreateActivity(e: Event) {
		e.preventDefault();
		if (!newActivityName.trim()) return;
		recentLogs = [
			{
				activity: `Activity "${newActivityName}" published`,
				type: 'Activity Created',
				performedBy: 'Super Admin',
				date: 'Jun 28, 2026',
				status: 'Completed'
			},
			...recentLogs
		];
		triggerToast(`Activity "${newActivityName}" published successfully!`);
		newActivityName = '';
		isCreateActivityModalOpen = false;
	}

	function handleCreateTrack(e: Event) {
		e.preventDefault();
		if (!newTrackName.trim()) return;
		recentLogs = [
			{
				activity: `Track "${newTrackName}" updated`,
				type: 'Track Updated',
				performedBy: 'Super Admin',
				date: 'Jun 28, 2026',
				status: 'Completed'
			},
			...recentLogs
		];
		triggerToast(`Track "${newTrackName}" created successfully!`);
		newTrackName = '';
		isCreateTrackModalOpen = false;
	}

	function handleGenerateReport() {
		triggerToast('System activity report generation initiated. Downloading...');
		isGenerateReportModalOpen = false;
	}

	// ── Reports Center (Step 7) ─────────────────────────────────────────────────
	type ReportStatus = 'Ready' | 'Processing' | 'Failed';
	type ReportFormat = 'PDF' | 'Excel' | 'CSV';

	interface GeneratedReport {
		id: number;
		name: string;
		by: string;
		date: string;
		format: ReportFormat;
		status: ReportStatus;
		fileName: string;
		// Original generation parameters, kept so a failed report can be retried.
		type: string;
		course: string;
		semester: string;
		dateFrom: string;
		dateTo: string;
	}

	const reportsBase = `${API_BASE_URL}/api/admin/platform/reports`;

	// Generate form state
	const reportTypes = [
		'Student Performance',
		'Mentor Analytics',
		'Activity Participation',
		'Credit Distribution',
		'Department Analytics',
		'Certificate Verification',
		'Leaderboard Rankings',
		'Semester Summary'
	];
	// Course and semester options are loaded from the API (canonical student data)
	// so they never drift from the seeded/registered course names. "All …" is
	// prepended as the no-filter choice.
	let reportCourses = $state<string[]>(['All Courses']);
	let reportSemesters = $state<string[]>(['All Semesters']);

	let reportType = $state('Student Performance');
	let reportCourse = $state('All Courses');
	let reportSemester = $state('All Semesters');
	let reportFrom = $state('2025-08-01');
	let reportTo = $state('2026-05-31');
	let reportFormat = $state<ReportFormat>('CSV');

	// Cadence chosen when scheduling the currently-configured report.
	const scheduleFrequencies = ['Daily', 'Weekly', 'Monthly', 'Quarterly'];
	let scheduleFrequency = $state('Monthly');

	// Which filters a given report type actually honours. Filters that do not
	// apply are disabled in the form so the UI never implies a filter is in
	// effect when the backend ignores it. Mentor Analytics is a per-admin batch
	// rollup and honours none of the student-scoped filters.
	const reportFilterSupport: Record<string, { course: boolean; semester: boolean; date: boolean }> =
		{
			'Mentor Analytics': { course: false, semester: false, date: false }
		};
	function filterApplies(type: string, filter: 'course' | 'semester' | 'date'): boolean {
		return reportFilterSupport[type]?.[filter] ?? true;
	}

	// Reports Center overview figures, loaded from the API.
	let reportSummary = $state({
		total_reports: 0,
		scheduled_reports: 0,
		monthly_downloads: 0,
		storage_bytes: 0
	});

	// Generated reports, loaded from the API.
	let recentReports = $state<GeneratedReport[]>([]);

	// Quick templates, loaded from the API. `type` maps a template to a report type.
	let quickTemplates = $state<{ name: string; desc: string; type: string }[]>([]);

	// Scheduled reports, loaded from the API.
	interface ScheduledReportItem {
		id: number;
		name: string;
		freq: string;
		format: string;
		type: string;
		enabled: boolean;
	}
	let scheduledReports = $state<ScheduledReportItem[]>([]);

	// Export center options. Names/descriptions are fixed; `type` drives the export
	// endpoint and `count` is refreshed from the API.
	let exportOptions = $state([
		{ name: 'Export Student Database', desc: 'Full student records', type: 'students', count: '—' },
		{
			name: 'Export Activity Records',
			desc: 'All logged activities',
			type: 'activities',
			count: '—'
		},
		{
			name: 'Export Credit Reports',
			desc: 'Credit distribution data',
			type: 'credits',
			count: '—'
		},
		{
			name: 'Export Certificate Data',
			desc: 'Verified certificates',
			type: 'certificates',
			count: '—'
		}
	]);

	// Report activity audit log, loaded from the API.
	let reportAuditLog = $state<{ action: string; type: string; user: string; time: string }[]>([]);

	// Institutional overview, loaded from the API.
	let institutionalRaw = $state({
		registered_students: 0,
		active_students: 0,
		activities_conducted: 0,
		avg_credits: 0
	});

	// Stat cards keep their fixed labels/colours; only the values come from the API.
	const institutionalStats = $derived([
		{
			value: institutionalRaw.registered_students.toLocaleString(),
			label: 'Registered Students',
			color: 'text-indigo-600'
		},
		{
			value: institutionalRaw.active_students.toLocaleString(),
			label: 'Active Students',
			color: 'text-emerald-600'
		},
		{
			value: institutionalRaw.activities_conducted.toLocaleString(),
			label: 'Activities Conducted',
			color: 'text-purple-600'
		},
		{
			value: institutionalRaw.avg_credits.toFixed(1),
			label: 'Avg Credits Earned',
			color: 'text-amber-600'
		}
	]);

	// Avg earned credits per student, per department. The chart's y-axis runs 0–100,
	// which matches the 100-credit graduation target, so scores map directly.
	let deptScores = $state<{ dept: string; score: number }[]>([]);

	// Monthly participation trend (dual line), loaded from the API.
	let participationMonths = $state<string[]>([]);
	let activitiesTrend = $state<number[]>([]);
	let certificatesTrend = $state<number[]>([]);

	// The trend chart scales to the data: the y-axis top is the largest value seen
	// (floored at 4 so a near-empty chart still reads sensibly).
	const trendMin = 0;
	const trendMax = $derived(Math.max(4, ...activitiesTrend, ...certificatesTrend));
	const trendAxisLabels = $derived([1, 0.75, 0.5, 0.25].map((f) => Math.round(trendMax * f)));

	function trendLine(data: number[]): string {
		return data
			.map((v, i) => {
				const x = data.length > 1 ? (i / (data.length - 1)) * 100 : 0;
				const y = 40 - ((v - trendMin) / (trendMax - trendMin)) * 36 - 2;
				return `${x},${y}`;
			})
			.join(' ');
	}
	const activitiesPoints = $derived(trendLine(activitiesTrend));
	const certificatesPoints = $derived(trendLine(certificatesTrend));

	// ── Reports Center: API wiring ──────────────────────────────────────────────

	// unauthorized bounces the user back to sign-in on a 401 and reports whether
	// the caller should stop. It mirrors the check used across the dashboard.
	async function unauthorized(res: Response): Promise<boolean> {
		if (res.status === 401) {
			localStorage.removeItem('superadmin_token');
			await goto('/super-admin-portal');
			return true;
		}
		return false;
	}

	// triggerBlobDownload saves a fetched blob to the user's device. Report files
	// sit behind an auth header, so they are fetched then downloaded rather than
	// linked directly.
	function triggerBlobDownload(blob: Blob, filename: string) {
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement('a');
		anchor.href = url;
		anchor.download = filename;
		document.body.appendChild(anchor);
		anchor.click();
		anchor.remove();
		URL.revokeObjectURL(url);
	}

	function formatReportDate(iso: string): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatAuditTime(iso: string): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	// mapApiReport converts an API report into the shape the table renders, keeping
	// the original parameters so a failed report can be retried.
	function mapApiReport(r: {
		id: number;
		name: string;
		type: string;
		course: string;
		semester: string;
		date_from: string | null;
		date_to: string | null;
		format: ReportFormat;
		status: ReportStatus;
		file_name: string;
		generated_by: string;
		created_at: string;
	}): GeneratedReport {
		return {
			id: r.id,
			name: r.name,
			by: r.generated_by || 'Super Admin',
			date: formatReportDate(r.created_at),
			format: r.format,
			status: r.status,
			fileName: r.file_name ?? '',
			type: r.type,
			course: r.course ?? '',
			semester: r.semester ?? '',
			dateFrom: r.date_from ? r.date_from.slice(0, 10) : '',
			dateTo: r.date_to ? r.date_to.slice(0, 10) : ''
		};
	}

	async function loadReportsData() {
		try {
			const [
				summaryRes,
				reportsRes,
				scheduledRes,
				countsRes,
				auditRes,
				templatesRes,
				institutionalRes,
				filtersRes
			] = await Promise.all([
				fetch(`${reportsBase}/summary`, { headers: authHeaders() }),
				fetch(`${reportsBase}`, { headers: authHeaders() }),
				fetch(`${reportsBase}/scheduled`, { headers: authHeaders() }),
				fetch(`${reportsBase}/export/counts`, { headers: authHeaders() }),
				fetch(`${reportsBase}/audit`, { headers: authHeaders() }),
				fetch(`${reportsBase}/templates`, { headers: authHeaders() }),
				fetch(`${reportsBase}/institutional`, { headers: authHeaders() }),
				fetch(`${reportsBase}/filters`, { headers: authHeaders() })
			]);

			if (
				[
					summaryRes,
					reportsRes,
					scheduledRes,
					countsRes,
					auditRes,
					templatesRes,
					institutionalRes,
					filtersRes
				].some((res) => res.status === 401)
			) {
				localStorage.removeItem('superadmin_token');
				await goto('/super-admin-portal');
				return;
			}

			if (summaryRes.ok) reportSummary = await summaryRes.json();

			if (reportsRes.ok) {
				const { reports } = await reportsRes.json();
				recentReports = (reports ?? []).map(mapApiReport);
			}

			if (scheduledRes.ok) {
				const { scheduled } = await scheduledRes.json();
				scheduledReports = (scheduled ?? []).map(
					(s: {
						id: number;
						name: string;
						frequency: string;
						format: string;
						type: string;
						enabled: boolean;
					}) => ({
						id: s.id,
						name: s.name,
						freq: s.frequency,
						format: s.format,
						type: s.type,
						enabled: s.enabled
					})
				);
			}

			if (countsRes.ok) {
				const counts: Record<string, number> = await countsRes.json();
				exportOptions = exportOptions.map((opt) => ({
					...opt,
					count: `${(counts[opt.type] ?? 0).toLocaleString()} records`
				}));
			}

			if (auditRes.ok) {
				const { logs } = await auditRes.json();
				reportAuditLog = (logs ?? []).map(
					(log: {
						action: string;
						category: string;
						type: string;
						user: string;
						created_at: string;
					}) => ({
						action: log.action,
						type: log.type || log.category,
						user: log.user || 'Super Admin',
						time: formatAuditTime(log.created_at)
					})
				);
			}

			if (templatesRes.ok) {
				const { templates } = await templatesRes.json();
				quickTemplates = (templates ?? []).map(
					(tpl: { name: string; description: string; type: string }) => ({
						name: tpl.name,
						desc: tpl.description,
						type: tpl.type
					})
				);
			}

			if (institutionalRes.ok) {
				const { stats, department_scores, trend } = await institutionalRes.json();
				institutionalRaw = stats ?? institutionalRaw;
				deptScores = department_scores ?? [];
				participationMonths = trend?.months ?? [];
				activitiesTrend = trend?.activities ?? [];
				certificatesTrend = trend?.certificates ?? [];
			}

			if (filtersRes.ok) {
				const { courses, semesters } = await filtersRes.json();
				reportCourses = ['All Courses', ...(courses ?? [])];
				reportSemesters = [
					'All Semesters',
					...(semesters ?? []).map((s: number) => `Semester ${s}`)
				];
				// Keep the current selection valid if the loaded options dropped it.
				if (!reportCourses.includes(reportCourse)) reportCourse = 'All Courses';
				if (!reportSemesters.includes(reportSemester)) reportSemester = 'All Semesters';
			}
		} catch {
			loadError = 'Could not load reports data. Please try again.';
		}
	}

	// Handlers
	async function runReportGeneration() {
		try {
			const res = await fetch(`${reportsBase}/generate`, {
				method: 'POST',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					type: reportType,
					course: reportCourse,
					semester: reportSemester,
					date_from: reportFrom,
					date_to: reportTo,
					format: reportFormat
				})
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to generate report');
				return;
			}

			triggerToast(`"${reportType} Report" generated successfully!`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to generate report');
		}
	}

	function resetReportForm() {
		reportType = 'Student Performance';
		reportCourse = 'All Courses';
		reportSemester = 'All Semesters';
		reportFrom = '2025-08-01';
		reportTo = '2026-05-31';
		reportFormat = 'CSV';
		triggerToast('Report filters reset.');
	}

	async function downloadReport(report: GeneratedReport) {
		try {
			const res = await fetch(`${reportsBase}/${report.id}/download`, { headers: authHeaders() });

			if (await unauthorized(res)) return;

			if (!res.ok) {
				triggerToast('Failed to download report');
				return;
			}

			// Use the server-provided filename (correct extension), not a forced .csv.
			const fallback = `${report.name}.csv`;
			triggerBlobDownload(await res.blob(), report.fileName || fallback);
			triggerToast(`Downloaded "${report.name}".`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to download report');
		}
	}

	// A failed report is retried by regenerating it with its original parameters.
	async function retryReport(index: number) {
		const report = recentReports[index];
		try {
			const res = await fetch(`${reportsBase}/generate`, {
				method: 'POST',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					type: report.type,
					course: report.course,
					semester: report.semester,
					date_from: report.dateFrom,
					date_to: report.dateTo,
					format: report.format
				})
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				triggerToast('Failed to retry report');
				return;
			}

			triggerToast(`Retried "${report.name}".`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to retry report');
		}
	}

	function useTemplate(tpl: { name: string; type: string }) {
		if (reportTypes.includes(tpl.type)) {
			reportType = tpl.type;
		}
		triggerToast(`Template "${tpl.name}" loaded into the generator.`);
	}

	// Schedules the report currently configured in the generate form on the
	// selected cadence (scheduleFrequency).
	async function scheduleNewReport() {
		try {
			const res = await fetch(`${reportsBase}/scheduled`, {
				method: 'POST',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: `${reportType} Report`,
					type: reportType,
					frequency: scheduleFrequency,
					format: reportFormat
				})
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to schedule report');
				return;
			}

			triggerToast(`Scheduled "${reportType} Report" (${scheduleFrequency}).`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to schedule report');
		}
	}

	// Enables or disables a scheduled report. A disabled schedule stays in the
	// list (so it can be re-enabled) but no longer runs.
	async function toggleSchedule(sched: ScheduledReportItem) {
		try {
			const res = await fetch(`${reportsBase}/scheduled/${sched.id}`, {
				method: 'PUT',
				headers: { ...authHeaders(), 'Content-Type': 'application/json' },
				body: JSON.stringify({ enabled: !sched.enabled })
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to update schedule');
				return;
			}

			triggerToast(`"${sched.name}" ${sched.enabled ? 'disabled' : 'enabled'}.`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to update schedule');
		}
	}

	// Permanently removes a scheduled report.
	async function deleteSchedule(sched: ScheduledReportItem) {
		try {
			const res = await fetch(`${reportsBase}/scheduled/${sched.id}`, {
				method: 'DELETE',
				headers: authHeaders()
			});

			if (await unauthorized(res)) return;

			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				triggerToast(data.error ?? 'Failed to delete schedule');
				return;
			}

			triggerToast(`Deleted schedule "${sched.name}".`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to delete schedule');
		}
	}

	async function runExport(opt: { name: string; type: string }) {
		try {
			const res = await fetch(`${reportsBase}/export?type=${opt.type}`, { headers: authHeaders() });

			if (await unauthorized(res)) return;

			if (!res.ok) {
				triggerToast('Failed to export data');
				return;
			}

			triggerBlobDownload(await res.blob(), `${opt.type}_export.csv`);
			triggerToast(`${opt.name} ready. Downloading…`);
			await loadReportsData();
		} catch {
			triggerToast('Failed to export data');
		}
	}

	function reportStatusClass(status: ReportStatus | string): string {
		switch (status) {
			case 'Ready':
			case 'Success':
				return 'bg-emerald-50 text-emerald-700 border-emerald-100';
			case 'Processing':
			case 'Scheduled':
				return 'bg-amber-50 text-amber-700 border-amber-100';
			default:
				return 'bg-rose-50 text-rose-700 border-rose-100';
		}
	}

	function reportStatusDot(status: ReportStatus | string): string {
		switch (status) {
			case 'Ready':
			case 'Success':
				return 'bg-emerald-500';
			case 'Processing':
			case 'Scheduled':
				return 'bg-amber-500';
			default:
				return 'bg-rose-500';
		}
	}

	function toggleMobileSidebar() {
		isMobileSidebarOpen = !isMobileSidebarOpen;
	}

	function handleLogout() {
		localStorage.removeItem('superadmin_token');
		goto('/super-admin-portal');
	}
</script>

<svelte:head>
	<title>Super Admin Console | iSPARC</title>
	<meta
		name="description"
		content="iSPARC Super Admin Portal. Configure tracks, view activity logs, manage administration settings, and oversee platform status."
	/>
</svelte:head>

<div class="min-h-screen bg-[#F7F6F3] text-slate-800 flex font-sans animate-fade-in">
	<!-- ==================== SIDEBAR ==================== -->
	<!-- Desktop Sidebar -->
	<aside
		class="hidden lg:flex flex-col w-64 bg-white border-r border-slate-200 h-screen sticky top-0 shrink-0 select-none"
	>
		<!-- Logo area -->
		<div class="h-[72px] flex items-center px-6 gap-3.5 border-b border-slate-100">
			<!-- Red square badge with white SA -->
			<div
				class="w-9 h-9 bg-[#881B1B] rounded-lg flex items-center justify-center text-white font-bold font-serif text-[11px] shrink-0"
			>
				SA
			</div>
			<div class="flex flex-col">
				<span class="text-xl font-bold tracking-tight text-slate-900 font-serif">
					i<span class="text-[#881B1B]">SPARC</span>
				</span>
				<span class="text-[9px] font-bold text-slate-600 tracking-wider uppercase -mt-1">
					IIPS DAVV CELL
				</span>
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
						: 'text-slate-800 hover:bg-slate-50 hover:text-slate-900'}"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						stroke-width="2"
						class="w-5 h-5 shrink-0"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
					</svg>
					<span class="text-left">
						{item.name}
					</span>
				</button>
			{/each}
		</nav>

		<!-- Logout -->
		<div class="p-4 border-t border-slate-100">
			<button
				onclick={handleLogout}
				class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-800 hover:bg-rose-50 hover:text-rose-600 transition-colors"
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
						class="w-9 h-9 bg-[#881B1B] rounded-lg flex items-center justify-center text-white font-bold font-serif text-[11px] shrink-0"
					>
						SA
					</div>
					<div class="flex flex-col">
						<span class="text-xl font-bold tracking-tight text-slate-900 font-serif">
							i<span class="text-[#881B1B]">SPARC</span>
						</span>
						<span class="text-[9px] font-bold text-slate-600 tracking-wider uppercase -mt-1">
							IIPS DAVV CELL
						</span>
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
							: 'text-slate-800 hover:bg-slate-50 hover:text-slate-900'}"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
							class="w-5 h-5 shrink-0"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
						</svg>
						<span class="text-left">
							{item.name}
							<span> </span></span
						></button
					>
				{/each}
			</nav>

			<div class="p-4 border-t border-slate-100">
				<button
					onclick={() => {
						handleLogout();
						toggleMobileSidebar();
					}}
					class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-[13px] font-bold text-slate-800 hover:bg-rose-50 hover:text-rose-600 transition-colors"
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
			class="bg-white border-b border-slate-200 h-[72px] flex items-center justify-between px-4 sm:px-6 lg:px-8 sticky top-0 z-30 select-none"
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
						{currentTab === 'Dashboard' ? 'Welcome Back, Super Admin!' : currentTab}
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
					<span class="absolute top-2 right-2.5 w-2 h-2 bg-[#881B1B] rounded-full"></span>
				</button>

				{#if isNotificationsOpen}
					<div
						transition:slide={{ duration: 150 }}
						class="absolute right-0 top-12 w-80 bg-white border border-slate-200 rounded-xl shadow-lg py-2.5 z-40 mt-1 animate-fade-in"
					>
						<h3
							class="px-4 py-1.5 text-xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-100"
						>
							System Alerts
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
							{/each}
						</div>
					</div>
				{/if}

				<!-- Profile Pill -->
				<div class="flex items-center gap-3 pl-2 border-l border-slate-200 select-none">
					<div
						class="w-10 h-10 rounded-full bg-[#881B1B] text-white flex items-center justify-center font-bold text-sm border-2 border-white shadow-sm shrink-0 font-serif"
					>
						SA
					</div>
					<div class="hidden sm:flex flex-col">
						<span class="text-xs font-bold text-slate-900 leading-none">Super Admin</span>
						<span class="text-[9px] text-slate-400 font-bold tracking-wider block mt-1 uppercase">
							Platform Administrator
						</span>
					</div>
				</div>
			</div>
		</header>

		<!-- Main Content Body -->
		<main class="flex-grow p-4 sm:p-6 lg:p-8 space-y-6 overflow-y-auto max-w-7xl mx-auto w-full">
			{#if loadError}
				<div
					class="flex items-center justify-between gap-4 p-4 bg-red-50 border border-red-200 rounded-xl"
					role="alert"
				>
					<span class="text-xs font-semibold text-red-700">{loadError}</span>
					<button
						type="button"
						onclick={loadPlatformData}
						class="px-3 py-1.5 text-[11px] font-bold uppercase tracking-wider text-red-700 border border-red-200 rounded-lg hover:bg-red-100 transition-colors"
					>
						Retry
					</button>
				</div>
			{/if}

			{#if currentTab === 'Dashboard'}
				<!-- Dashboard Statistics Cards Grid -->
				<section
					class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6"
					aria-label="Super Admin metrics overview"
				>
					<!-- Card 1: Total Students -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{stats.total_students}</span
							>
							<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100">
								<!-- Academic / Student icon -->
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
										d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Total Students
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Active student accounts
							</p>
						</div>
					</div>

					<!-- Card 2: Total Admins -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{stats.total_admins}</span>
							<div class="p-2.5 rounded-lg bg-blue-50 text-blue-600 border border-blue-100">
								<!-- Admin / Users icon -->
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
										d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"
									/>
									A
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Total Admins</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Active admin accounts
							</p>
						</div>
					</div>

					<!-- Card 3: Total Activities -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{stats.total_activities}</span
							>
							<div
								class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100"
							>
								<!-- Graph / Activity icon -->
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
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Total Activities
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Total campus activities
							</p>
						</div>
					</div>

					<!-- Card 4: Active Tracks -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{stats.active_tracks}</span
							>
							<div class="p-2.5 rounded-lg bg-purple-50 text-purple-600 border border-purple-100">
								<!-- Layers / Track icon -->
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
										d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L21.75 12l-4.179 2.25m0 0L21.75 16.5l-9.75 5.25-9.75-5.25L6.429 14.25m11.142 0L12 16.5m0-13.5L21.75 7.5 12 12.75 2.25 7.5 12 3z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Active Tracks
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Categorized activity
							</p>
						</div>
					</div>
				</section>
				<!-- Middle Grid section (Step 3 & 4) -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
					<!-- User Management Overview (lg:col-span-2) -->
					<div
						class="lg:col-span-2 bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden"
					>
						<div
							class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20"
						>
							<h3 class="text-sm font-bold font-serif text-slate-905">User Management Overview</h3>
							<button
								onclick={() => {
									currentTab = 'User Management';
									userFilter = 'All';
									userSearch = '';
								}}
								class="text-[#881B1B] hover:underline text-xs font-bold uppercase tracking-wider"
							>
								View All
							</button>
						</div>
						<div class="overflow-x-auto">
							<table class="w-full text-left border-collapse">
								<thead>
									<tr
										class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
									>
										<th class="py-3.5 px-5">User Name</th>
										<th class="py-3.5 px-5">Role</th>
										<th class="py-3.5 px-5">Department</th>
										<th class="py-3.5 px-5">Status</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-100 text-xs font-sans">
									{#each recentUsers as user}
										<tr class="hover:bg-slate-50/30 transition-colors">
											<td class="py-4 px-5">
												<div class="flex flex-col">
													<span class="font-bold text-slate-800">{user.name}</span>
													<span class="text-[10px] text-slate-400 font-semibold mt-0.5 select-all"
														>{user.id}</span
													>
												</div>
											</td>
											<td class="py-4 px-5">
												<span
													class="inline-flex items-center px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border
													{user.role === 'Student'
														? 'bg-slate-50 text-slate-650 border-slate-200'
														: 'bg-blue-50 text-blue-700 border-blue-100'}"
												>
													{user.role}
												</span>
											</td>
											<td class="py-4 px-5 text-slate-500 font-semibold">{user.dept}</td>
											<td class="py-4 px-5">
												<span
													class="inline-flex items-center gap-1.5 font-bold
													{user.status === 'Active' ? 'text-emerald-600' : 'text-amber-600'}"
												>
													<span
														class="w-1.5 h-1.5 rounded-full shrink-0
														{user.status === 'Active' ? 'bg-emerald-600' : 'bg-amber-500 animate-pulse'}"
													></span>
													{user.status}
												</span>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>

					<!-- Quick Actions Panel (Step 4) -->
					<div class="bg-white border border-slate-200 rounded-xl p-5 shadow-xs space-y-4">
						<h3 class="text-sm font-bold font-serif text-slate-905">Quick Actions</h3>
						<div class="h-px bg-slate-100 my-2"></div>

						<div class="grid grid-cols-2 gap-4 select-none">
							<!-- Button 1: Create User (Solid Red) -->
							<button
								onclick={() => (isCreateUserModalOpen = true)}
								class="p-5 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white rounded-[20px] shadow-xs flex flex-col items-center justify-center text-center space-y-2.5 h-[160px] w-full transition duration-200 focus:outline-none"
							>
								<!-- User plus icon -->
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="2"
									stroke="currentColor"
									class="w-6 h-6 text-white"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M19 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM4 19.235A8.902 8.902 0 0110 18a8.902 8.902 0 016 1.235c.19.115.3.322.3.54v.725c0 .19-.153.344-.344.344H4.344A.344.344 0 014 20.5v-.725c0-.218.11-.425.3-.54z"
									/>
								</svg>
								<div class="space-y-1">
									<span class="font-extrabold text-xs block font-sans">Create User</span>
									<span class="text-[9px] text-white/80 font-medium block font-sans leading-tight"
										>Add student or admin</span
									>
								</div>
							</button>

							<!-- Button 2: Create Activity -->
							<button
								onclick={() => {
									currentTab = 'Activity Management';
									isManageActivityModalOpen = true;
								}}
								class="p-5 bg-[#EFECE9] hover:bg-[#E4DFDB] text-slate-800 rounded-[20px] flex flex-col items-center justify-center text-center space-y-2.5 h-[160px] w-full transition duration-200 focus:outline-none"
							>
								<!-- Activity graph icon -->
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="2"
									stroke="currentColor"
									class="w-6 h-6 text-slate-700"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M2.25 18 9 11.25l4.306 4.306a11.95 11.95 0 015.814-5.518l2.74-1.22m0 0-5.94-2.281m5.94 2.28-2.28 5.941"
									/>
								</svg>
								<div class="space-y-1">
									<span class="font-extrabold text-xs block font-sans text-slate-905"
										>Create Activity</span
									>
									<span
										class="text-[9px] text-slate-500 font-semibold block font-sans leading-tight"
										>Publish new activity</span
									>
								</div>
							</button>

							<!-- Button 3: Create Track -->
							<button
								onclick={() => (isCreateTrackModalOpen = true)}
								class="p-5 bg-[#EFECE9] hover:bg-[#E4DFDB] text-slate-800 rounded-[20px] flex flex-col items-center justify-center text-center space-y-2.5 h-[160px] w-full transition duration-200 focus:outline-none"
							>
								<!-- Track icon -->
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="2"
									stroke="currentColor"
									class="w-6 h-6 text-slate-700"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L21.75 12l-4.179 2.25m0 0L21.75 16.5l-9.75 5.25-9.75-5.25L6.429 14.25m11.142 0L12 16.5m0-13.5L21.75 7.5 12 12.75 2.25 7.5 12 3z"
									/>
								</svg>
								<div class="space-y-1">
									<span class="font-extrabold text-xs block font-sans text-slate-905"
										>Create Track</span
									>
									<span
										class="text-[9px] text-slate-500 font-semibold block font-sans leading-tight"
										>Set up a new track</span
									>
								</div>
							</button>

							<!-- Button 4: Generate Report -->
							<button
								onclick={() => (isGenerateReportModalOpen = true)}
								class="p-5 bg-[#EFECE9] hover:bg-[#E4DFDB] text-slate-800 rounded-[20px] flex flex-col items-center justify-center text-center space-y-2.5 h-[160px] w-full transition duration-200 focus:outline-none"
							>
								<!-- Document download icon -->
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="2"
									stroke="currentColor"
									class="w-6 h-6 text-slate-700"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
									/>
								</svg>
								<div class="space-y-1">
									<span class="font-extrabold text-xs block font-sans text-slate-905"
										>Generate Report</span
									>
									<span
										class="text-[9px] text-slate-500 font-semibold block font-sans leading-tight"
										>Download activity data</span
									>
								</div>
							</button>
						</div>
					</div>
				</div>

				<!-- Recent System Activities Table Section (Step 5) -->
				<section class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
					<div
						class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20 select-none"
					>
						<div class="flex items-center gap-2">
							<h3 class="text-sm font-bold font-serif text-slate-905">Recent System Activities</h3>
							<span
								class="px-2 py-0.5 bg-rose-50 text-rose-705 font-extrabold text-[10px] uppercase rounded-md border border-rose-100"
							>
								Live Feed
							</span>
						</div>
						<button
							onclick={() => triggerToast('Navigating to full audit log...')}
							class="text-[#881B1B] hover:underline text-xs font-bold uppercase tracking-wider"
						>
							View Full Log
						</button>
					</div>

					<div class="overflow-x-auto">
						<table class="w-full text-left border-collapse">
							<thead>
								<tr
									class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
								>
									<th class="py-3.5 px-5">Activity</th>
									<th class="py-3.5 px-5">Type</th>
									<th class="py-3.5 px-5">Performed By</th>
									<th class="py-3.5 px-5">Date</th>
									<th class="py-3.5 px-5">Status</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-100 text-xs font-sans">
								{#each recentLogs as log}
									<tr class="hover:bg-slate-50/30 transition-colors">
										<td class="py-4 px-5 font-bold text-slate-800">{log.activity}</td>
										<td class="py-4 px-5">
											<span
												class="inline-flex items-center px-2.5 py-0.5 text-[9px] font-bold uppercase rounded-md border
												{log.type === 'User Added'
													? 'bg-slate-100 text-slate-650 border-slate-200'
													: log.type === 'Activity Created'
														? 'bg-blue-50 text-blue-700 border-blue-100'
														: log.type === 'Track Updated'
															? 'bg-purple-50 text-purple-700 border-purple-100'
															: log.type === 'Announcement'
																? 'bg-amber-50 text-amber-705 border-amber-100'
																: 'bg-emerald-50 text-emerald-700 border-emerald-100'}"
											>
												{log.type}
											</span>
										</td>
										<td class="py-4 px-5 text-slate-600 font-semibold">{log.performedBy}</td>
										<td class="py-4 px-5 text-slate-500 font-semibold">{log.date}</td>
										<td class="py-4 px-5">
											<span class="inline-flex items-center gap-1.5 font-bold text-emerald-600">
												<span class="w-1.5 h-1.5 rounded-full bg-emerald-600 shrink-0"></span>
												{log.status}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					<div
						class="p-4 border-t border-slate-100 bg-slate-50/30 text-slate-500 font-semibold text-[11px] select-none"
					>
						<span>Showing {recentLogs.length} of {recentLogs.length} recent system activities</span>
					</div>
				</section>
			{:else if currentTab === 'Activity Management'}
				<ActivityManagementView bind:isActivityModalOpen={isManageActivityModalOpen} />
			{:else if currentTab === 'Track Management'}
				<TrackManagementView />
			{:else if currentTab === 'Announcement Management'}
				<AnnouncementManagementView />
			{:else if currentTab === 'User Management'}
				<!-- User Statistics Cards Grid (Step 2) -->
				<section
					class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
					aria-label="User metrics overview"
				>
					<!-- Card 1: Total Students -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{totalStudentsCount}</span>
							<div class="p-2.5 rounded-lg bg-rose-50 text-rose-600 border border-rose-100">
								<!-- Book open icon -->
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
										d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Total Students
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Active Student Accounts
							</p>
						</div>
					</div>

					<!-- Card 2: Total Admins -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{totalAdminsCount}</span>
							<div class="p-2.5 rounded-lg bg-blue-50 text-blue-600 border border-blue-100">
								<!-- Users icon -->
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
										d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Total Admins</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Active Admin Accounts
							</p>
						</div>
					</div>

					<!-- Card 3: Pending Verification -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{pendingUsersCount}</span>
							<div class="p-2.5 rounded-lg bg-red-50 text-red-655 border border-red-100">
								<!-- Ribbon icon -->
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
										d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Pending Accounts
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Awaiting Verification
							</p>
						</div>
					</div>

					<!-- Card 4: Active Users -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">{activeUsersCount}</span>
							<div class="p-2.5 rounded-lg bg-[#E8F5E9] text-[#2E7D32] border border-[#C8E6C9]">
								<!-- Graph trending up icon -->
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
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Active Users</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								{activeRate}% active rate
							</p>
						</div>
					</div>
				</section>

				<!-- Main 2-Column Grid (Step 3 & 4 Layout) -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
					<!-- Left Column: User Management Table (lg:col-span-2) -->
					<div
						class="lg:col-span-2 bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden flex flex-col"
					>
						<!-- Table Header -->
						<div
							class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20 select-none"
						>
							<h3 class="text-sm font-bold font-serif text-slate-905">User Management Overview</h3>
							<button
								type="button"
								onclick={() => {
									userFilter = 'All';
									userSearch = '';
								}}
								class="text-[#881B1B] hover:underline text-xs font-bold uppercase tracking-wider"
							>
								View All
							</button>
						</div>

						<!-- Table Filters & Search Controls -->
						<div
							class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white select-none"
						>
							<!-- Filter Tabs -->
							<div class="flex flex-wrap gap-1.5">
								{#each ['All', 'Student', 'Admin'] as roleType}
									<button
										type="button"
										onclick={() => (userFilter = roleType)}
										class="px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all
											{userFilter === roleType
											? 'bg-[#881B1B] text-white shadow-xs'
											: 'bg-slate-50 text-slate-500 hover:bg-slate-100'}"
									>
										{roleType === 'All'
											? 'All Users'
											: roleType === 'Student'
												? 'Students'
												: 'Admins'}
									</button>
								{/each}
							</div>

							<!-- Search Bar -->
							<div class="relative w-full sm:w-64">
								<input
									type="text"
									bind:value={userSearch}
									placeholder="Search by name, roll no., or email..."
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

						<!-- Table Grid -->
						<div class="overflow-x-auto">
							<table class="w-full text-left border-collapse">
								<thead>
									<tr
										class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider select-none"
									>
										<th class="py-3.5 px-5">User Name</th>
										<th class="py-3.5 px-5">Role</th>
										<th class="py-3.5 px-5">Department</th>
										<th class="py-3.5 px-5">Status</th>
										<th class="py-3.5 px-5 text-right sm:text-left">Actions</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-100 text-xs font-sans">
									{#if filteredUsers.length === 0}
										<tr>
											<td
												colspan="5"
												class="py-8 text-center text-slate-400 font-semibold select-none"
											>
												No users found matching search filters.
											</td>
										</tr>
									{:else}
										{#each filteredUsers as user}
											<tr class="hover:bg-slate-50/30 transition-colors">
												<!-- User Info (Initials Avatar + Name & ID) -->
												<td class="py-4 px-5">
													<div class="flex items-center gap-3">
														<div
															class="w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs uppercase shrink-0 select-none {user.avatarBg}"
														>
															{user.name
																.split(' ')
																.map((n) => n[0])
																.join('')
																.slice(0, 2)}
														</div>
														<div class="flex flex-col">
															<span class="font-bold text-slate-800">{user.name}</span>
															<span
																class="text-[9px] text-slate-400 font-semibold mt-0.5 select-all"
																>{user.id}</span
															>
														</div>
													</div>
												</td>

												<!-- Role Badge -->
												<td class="py-4 px-5 select-none">
													<span
														class="inline-flex items-center px-2 py-0.5 text-[9px] font-bold uppercase rounded-full border
														{user.displayRole === 'Student'
															? 'bg-blue-50 text-blue-700 border-blue-100'
															: 'bg-rose-50 text-rose-700 border-rose-100'}"
													>
														{user.displayRole}
													</span>
												</td>

												<!-- Department -->
												<td class="py-4 px-5 text-slate-500 font-semibold">{user.dept}</td>

												<!-- Status Dot -->
												<td class="py-4 px-5 select-none">
													<span
														class="inline-flex items-center gap-1.5 font-bold
														{user.status === 'Active'
															? 'text-emerald-600'
															: user.status === 'Pending'
																? 'text-amber-600'
																: 'text-slate-400'}"
													>
														<span
															class="w-1.5 h-1.5 rounded-full shrink-0
															{user.status === 'Active'
																? 'bg-emerald-600'
																: user.status === 'Pending'
																	? 'bg-amber-500 animate-pulse'
																	: 'bg-slate-400'}"
														></span>
														{user.status}
													</span>
												</td>

												<!-- Action Buttons -->
												<td class="py-4 px-5">
													<div class="flex items-center gap-3">
														<button
															type="button"
															onclick={() => openViewUser(user)}
															aria-label="View user profile"
															class="text-blue-500 hover:text-blue-700 transition-colors p-0.5"
														>
															<!-- Eye icon -->
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
															onclick={() => openEditUser(user)}
															aria-label="Edit user details"
															class="text-amber-500 hover:text-amber-700 transition-colors p-0.5"
														>
															<!-- Pencil edit icon -->
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
															onclick={() => handleDeleteUser(user)}
															aria-label="Delete user"
															class="text-rose-500 hover:text-rose-700 transition-colors p-0.5"
														>
															<!-- Trash delete icon -->
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

						<!-- Table Footer Summary -->
						<div
							class="p-4 border-t border-slate-100 bg-slate-50/30 text-slate-500 font-semibold text-[11px] select-none"
						>
							<span>Showing {filteredUsers.length} of {userRegistry.length} users</span>
						</div>
					</div>

					<!-- Right Column Sidebar Panels (Step 4 Cards) -->
					<div class="space-y-6">
						<!-- Quick Actions Card -->
						<div class="bg-white border border-slate-200 rounded-xl p-5 shadow-xs space-y-4">
							<h3 class="text-sm font-bold font-serif text-slate-905">Quick Actions</h3>
							<div class="h-px bg-slate-100 my-2"></div>

							<div class="grid grid-cols-2 gap-4 select-none">
								<!-- Button 1: Add User (Solid Red) -->
								<button
									type="button"
									onclick={() => (isCreateUserModalOpen = true)}
									class="p-4 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white rounded-2xl shadow-xs flex flex-col items-center justify-center text-center space-y-2 h-[100px] w-full transition duration-200 focus:outline-none"
								>
									<!-- User plus icon -->
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										class="w-5 h-5 text-white"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M19 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM4 19.235A8.902 8.902 0 0110 18a8.902 8.902 0 016 1.235c.19.115.3.322.3.54v.725c0 .19-.153.344-.344.344H4.344A.344.344 0 014 20.5v-.725c0-.218.11-.425.3-.54z"
										/>
									</svg>
									<span class="font-extrabold text-[11px] block font-sans">Add User</span>
								</button>

								<!-- Button 2: Assign Mentor -->
								<button
									type="button"
									onclick={() => triggerToast('Assign Mentor flow coming soon...')}
									class="p-4 bg-[#F5F2F0] hover:bg-[#EAE5DF] text-slate-800 rounded-2xl flex flex-col items-center justify-center text-center space-y-2 h-[100px] w-full transition duration-200 focus:outline-none"
								>
									<!-- Link icon -->
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										class="w-5 h-5 text-[#881B1B]"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244"
										/>
									</svg>
									<span class="font-extrabold text-[11px] block font-sans text-slate-805"
										>Assign Mentor</span
									>
								</button>

								<!-- Button 3: Assign Role -->
								<button
									type="button"
									onclick={() => triggerToast('Assign Role flow coming soon...')}
									class="p-4 bg-[#F5F2F0] hover:bg-[#EAE5DF] text-slate-800 rounded-2xl flex flex-col items-center justify-center text-center space-y-2 h-[100px] w-full transition duration-200 focus:outline-none"
								>
									<!-- Shield / user cog icon -->
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										class="w-5 h-5 text-[#881B1B]"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
										/>
									</svg>
									<span class="font-extrabold text-[11px] block font-sans text-slate-805"
										>Assign Role</span
									>
								</button>

								<!-- Button 4: User Stats -->
								<button
									type="button"
									onclick={() => triggerToast('User statistics view selected.')}
									class="p-4 bg-[#F5F2F0] hover:bg-[#EAE5DF] text-slate-800 rounded-2xl flex flex-col items-center justify-center text-center space-y-2 h-[100px] w-full transition duration-200 focus:outline-none"
								>
									<!-- Speedometer / chart icon -->
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										class="w-5 h-5 text-[#881B1B]"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M10.5 6a7.5 7.5 0 107.5 7.5h-7.5V6z"
										/>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M13.5 10.5H21A7.5 7.5 0 0013.5 3v7.5z"
										/>
									</svg>
									<span class="font-extrabold text-[11px] block font-sans text-slate-805"
										>User Stats</span
									>
								</button>
							</div>
						</div>

						<!-- Account Status Summary Card -->
						<div
							class="bg-white border border-slate-200 rounded-xl p-5 shadow-xs space-y-4 select-none"
						>
							<h3 class="text-sm font-bold font-serif text-slate-905">Account Status</h3>
							<div class="h-px bg-slate-100 my-2"></div>

							<div class="space-y-3 font-sans text-xs">
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2 font-semibold text-slate-600">
										<span class="w-2 h-2 rounded-full bg-emerald-500"></span>
										Active
									</div>
									<span class="font-extrabold text-slate-800">{activeUsersCount}</span>
								</div>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2 font-semibold text-slate-600">
										<span class="w-2 h-2 rounded-full bg-amber-500"></span>
										Pending
									</div>
									<span class="font-extrabold text-[#C06014]">{pendingUsersCount}</span>
								</div>
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2 font-semibold text-slate-600">
										<span class="w-2 h-2 rounded-full bg-slate-400"></span>
										Inactive
									</div>
									<span class="font-extrabold text-slate-400">{inactiveUsersCount}</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			{:else if currentTab === 'Reports Center'}
				<!-- Reports Center (Step 7) -->
				<!-- Overview Stat Cards -->
				<section
					class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
					aria-label="Reports overview"
				>
					<!-- Total Reports -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{reportSummary.total_reports}</span
							>
							<div class="p-2.5 rounded-lg bg-blue-50 text-blue-600 border border-blue-100">
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
										d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Total Reports
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Generated all-time
							</p>
						</div>
					</div>

					<!-- Scheduled -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{reportSummary.scheduled_reports}</span
							>
							<div class="p-2.5 rounded-lg bg-amber-50 text-amber-600 border border-amber-100">
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
										d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Scheduled Reports
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Auto-generated
							</p>
						</div>
					</div>

					<!-- Downloads -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{reportSummary.monthly_downloads}</span
							>
							<div
								class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100"
							>
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
										d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">Downloads</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								This month
							</p>
						</div>
					</div>

					<!-- Last Generated -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{recentReports.length ? recentReports[0].date : '—'}</span
							>
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
										d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4 min-w-0">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Last Generated
							</h3>
							<p
								class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider truncate"
							>
								{recentReports.length ? recentReports[0].name : 'No reports yet'}
							</p>
						</div>
					</div>
				</section>

				<!-- Generate Report + Quick Templates -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
					<!-- Generate Report (col-span-2) -->
					<div
						class="lg:col-span-2 bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden"
					>
						<div class="p-5 border-b border-slate-100 bg-slate-50/20 select-none">
							<h3 class="text-sm font-bold font-serif text-slate-905">Generate Report</h3>
							<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
								Configure parameters and export your institutional report.
							</p>
						</div>
						<div class="p-5 space-y-5">
							<!-- Report Type -->
							<div class="flex flex-col gap-1.5">
								<label
									for="report-type"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider"
									>REPORT TYPE</label
								>
								<select
									id="report-type"
									bind:value={reportType}
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 bg-white focus:outline-none focus:border-slate-355"
								>
									{#each reportTypes as type}
										<option value={type}>{type}</option>
									{/each}
								</select>
								{#if !filterApplies(reportType, 'course') && !filterApplies(reportType, 'semester') && !filterApplies(reportType, 'date')}
									<p class="text-[10px] text-slate-400 font-semibold">
										This report is not scoped by course, semester or date range.
									</p>
								{/if}
							</div>

							<!-- Course / Semester -->
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div class="flex flex-col gap-1.5">
									<label
										for="report-course"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider"
										>COURSE / BATCH</label
									>
									<select
										id="report-course"
										bind:value={reportCourse}
										disabled={!filterApplies(reportType, 'course')}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 bg-white focus:outline-none focus:border-slate-355 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed"
									>
										{#each reportCourses as course}
											<option value={course}>{course}</option>
										{/each}
									</select>
								</div>
								<div class="flex flex-col gap-1.5">
									<label
										for="report-semester"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider">SEMESTER</label
									>
									<select
										id="report-semester"
										bind:value={reportSemester}
										disabled={!filterApplies(reportType, 'semester')}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs font-semibold text-slate-800 bg-white focus:outline-none focus:border-slate-355 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed"
									>
										{#each reportSemesters as sem}
											<option value={sem}>{sem}</option>
										{/each}
									</select>
								</div>
							</div>

							<!-- From / To -->
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div class="flex flex-col gap-1.5">
									<label
										for="report-from"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider">FROM</label
									>
									<input
										id="report-from"
										type="date"
										bind:value={reportFrom}
										disabled={!filterApplies(reportType, 'date')}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed"
									/>
								</div>
								<div class="flex flex-col gap-1.5">
									<label
										for="report-to"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider">TO</label
									>
									<input
										id="report-to"
										type="date"
										bind:value={reportTo}
										disabled={!filterApplies(reportType, 'date')}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed"
									/>
								</div>
							</div>

							<!-- Format: reports are exported as CSV -->
							<div class="flex flex-col gap-1.5">
								<span class="text-[10px] font-extrabold text-slate-650 tracking-wider">FORMAT</span>
								<span
									class="inline-flex w-fit items-center px-2.5 py-1 rounded-md text-[11px] font-bold bg-blue-50 text-blue-700 border border-blue-100"
								>
									CSV
								</span>
							</div>

							<!-- Actions -->
							<div class="flex items-center gap-3 pt-1">
								<button
									type="button"
									onclick={runReportGeneration}
									class="inline-flex items-center gap-1.5 px-4 py-2.5 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
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
											d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
										/>
									</svg>
									Generate Report
								</button>
								<button
									type="button"
									onclick={resetReportForm}
									class="px-4 py-2.5 border border-slate-200 hover:bg-slate-50 text-slate-600 font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
								>
									Reset
								</button>
							</div>
						</div>
					</div>

					<!-- Quick Templates (col-span-1) -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div class="p-5 border-b border-slate-100 bg-slate-50/20 select-none">
							<h3 class="text-sm font-bold font-serif text-slate-905">Quick Templates</h3>
							<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
								One-click report presets
							</p>
						</div>
						<div class="p-3 space-y-1 max-h-[420px] overflow-y-auto">
							{#each quickTemplates as tpl}
								<button
									type="button"
									onclick={() => useTemplate(tpl)}
									class="w-full flex items-center gap-3 p-2.5 rounded-lg hover:bg-slate-50 transition-colors text-left group"
								>
									<div
										class="w-8 h-8 rounded-lg bg-[#881B1B]/10 text-[#881B1B] flex items-center justify-center shrink-0 group-hover:bg-[#881B1B]/15 transition-colors"
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
												d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"
											/>
										</svg>
									</div>
									<div class="min-w-0 flex-grow">
										<span class="block text-xs font-bold text-slate-800 truncate">{tpl.name}</span>
										<span class="block text-[10px] text-slate-400 font-semibold truncate"
											>{tpl.desc}</span
										>
									</div>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2.5"
										stroke="currentColor"
										class="w-3.5 h-3.5 text-slate-300 group-hover:text-[#881B1B] transition-colors shrink-0"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="m8.25 4.5 7.5 7.5-7.5 7.5"
										/>
									</svg>
								</button>
							{/each}
						</div>
					</div>
				</div>

				<!-- Recent Reports -->
				<section class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
					<div
						class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20 select-none"
					>
						<div class="flex items-center gap-2">
							<h3 class="text-sm font-bold font-serif text-slate-905">Recent Reports</h3>
							<span
								class="px-2 py-0.5 bg-rose-50 text-rose-705 font-extrabold text-[10px] uppercase rounded-md border border-rose-100"
							>
								{recentReports.length} total
							</span>
						</div>
						<button
							onclick={() => triggerToast('Loading full report archive...')}
							class="text-[#881B1B] hover:underline text-xs font-bold uppercase tracking-wider"
						>
							View All Reports
						</button>
					</div>
					<div class="overflow-x-auto">
						<table class="w-full text-left border-collapse">
							<thead>
								<tr
									class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
								>
									<th class="py-3.5 px-5">Report Name</th>
									<th class="py-3.5 px-5">Generated By</th>
									<th class="py-3.5 px-5">Date</th>
									<th class="py-3.5 px-5">Status</th>
									<th class="py-3.5 px-5 text-center">Actions</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-100 text-xs font-sans">
								{#each recentReports as report, index (report.name + index)}
									<tr class="hover:bg-slate-50/30 transition-colors">
										<td class="py-4 px-5 font-bold text-slate-800">{report.name}</td>
										<td class="py-4 px-5 text-slate-600 font-semibold">{report.by}</td>
										<td class="py-4 px-5 text-slate-500 font-semibold whitespace-nowrap"
											>{report.date}</td
										>
										<td class="py-4 px-5">
											<span
												class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border {reportStatusClass(
													report.status
												)}"
											>
												<span
													class="w-1.5 h-1.5 rounded-full {reportStatusDot(
														report.status
													)} {report.status === 'Processing' ? 'animate-pulse' : ''}"
												></span>
												{report.status}
											</span>
										</td>
										<td class="py-4 px-5">
											<div class="flex items-center justify-center">
												{#if report.status === 'Ready'}
													<button
														type="button"
														onclick={() => downloadReport(report)}
														class="inline-flex items-center gap-1 px-2.5 py-1.5 text-[10px] font-bold text-white bg-[#881B1B] hover:bg-[#881B1B]/90 rounded-lg transition-colors"
													>
														<svg
															xmlns="http://www.w3.org/2000/svg"
															fill="none"
															viewBox="0 0 24 24"
															stroke-width="2"
															stroke="currentColor"
															class="w-3 h-3"
														>
															<path
																stroke-linecap="round"
																stroke-linejoin="round"
																d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
															/>
														</svg>
														Download
													</button>
												{:else if report.status === 'Processing'}
													<span class="text-[10px] font-bold text-amber-600 uppercase tracking-wide"
														>Processing…</span
													>
												{:else}
													<button
														type="button"
														onclick={() => retryReport(index)}
														class="inline-flex items-center gap-1 px-2.5 py-1.5 text-[10px] font-bold text-rose-600 border border-rose-200 bg-rose-50 hover:bg-rose-100 rounded-lg transition-colors"
													>
														<svg
															xmlns="http://www.w3.org/2000/svg"
															fill="none"
															viewBox="0 0 24 24"
															stroke-width="2"
															stroke="currentColor"
															class="w-3 h-3"
														>
															<path
																stroke-linecap="round"
																stroke-linejoin="round"
																d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"
															/>
														</svg>
														Retry
													</button>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</section>

				<!-- Scheduled Reports + Export Center -->
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
					<!-- Scheduled Reports -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div
							class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20 select-none"
						>
							<h3 class="text-sm font-bold font-serif text-slate-905">Scheduled Reports</h3>
							<div class="flex items-center gap-2">
								<select
									bind:value={scheduleFrequency}
									aria-label="Schedule frequency"
									class="px-2 py-1.5 border border-slate-200 rounded-lg text-[11px] font-semibold text-slate-700 bg-white focus:outline-none focus:border-slate-355"
								>
									{#each scheduleFrequencies as freq}
										<option value={freq}>{freq}</option>
									{/each}
								</select>
								<button
									type="button"
									onclick={scheduleNewReport}
									class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-[10px] uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										fill="none"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										class="w-3.5 h-3.5"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											d="M12 4.5v15m7.5-7.5h-15"
										/>
									</svg>
									Schedule New
								</button>
							</div>
						</div>
						<div class="p-3 space-y-1">
							{#each scheduledReports as sched}
								<div
									class="flex items-center gap-3 p-3 rounded-lg hover:bg-slate-50 transition-colors"
								>
									<div
										class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 border {sched.enabled
											? 'bg-amber-50 text-amber-600 border-amber-100'
											: 'bg-slate-50 text-slate-400 border-slate-200'}"
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
												d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
											/>
										</svg>
									</div>
									<div class="min-w-0 flex-grow">
										<span class="block text-xs font-bold text-slate-800 truncate">{sched.name}</span
										>
										<span class="block text-[10px] text-slate-400 font-semibold truncate"
											>{sched.freq}</span
										>
									</div>
									{#if sched.enabled}
										<span
											class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border bg-emerald-50 text-emerald-700 border-emerald-100 shrink-0"
										>
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
											Scheduled
										</span>
									{:else}
										<span
											class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border bg-slate-100 text-slate-500 border-slate-200 shrink-0"
										>
											<span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
											Disabled
										</span>
									{/if}
									<div class="flex items-center gap-1 shrink-0">
										<button
											type="button"
											onclick={() => toggleSchedule(sched)}
											title={sched.enabled ? 'Disable schedule' : 'Enable schedule'}
											aria-label={sched.enabled ? 'Disable schedule' : 'Enable schedule'}
											class="p-1.5 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors focus:outline-none"
										>
											{#if sched.enabled}
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
														d="M15.75 5.25v13.5m-7.5-13.5v13.5"
													/>
												</svg>
											{:else}
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
														d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 0 1 0 1.971l-11.54 6.347a1.125 1.125 0 0 1-1.667-.985V5.653Z"
													/>
												</svg>
											{/if}
										</button>
										<button
											type="button"
											onclick={() => deleteSchedule(sched)}
											title="Delete schedule"
											aria-label="Delete schedule"
											class="p-1.5 rounded-md text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition-colors focus:outline-none"
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
													d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
												/>
											</svg>
										</button>
									</div>
								</div>
							{:else}
								<p class="px-3 py-6 text-center text-[11px] text-slate-400 font-semibold">
									No scheduled reports yet.
								</p>
							{/each}
						</div>
					</div>

					<!-- Export Center -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div class="p-5 border-b border-slate-100 bg-slate-50/20 select-none">
							<h3 class="text-sm font-bold font-serif text-slate-905">Export Center</h3>
							<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
								Bulk export platform datasets
							</p>
						</div>
						<div class="p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
							{#each exportOptions as opt}
								<button
									type="button"
									onclick={() => runExport(opt)}
									class="p-4 rounded-xl border border-slate-150 hover:border-[#881B1B]/30 hover:bg-slate-50/60 transition-colors text-left flex flex-col gap-2 group"
								>
									<div
										class="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 border border-blue-100 flex items-center justify-center"
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
												d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
											/>
										</svg>
									</div>
									<div>
										<span class="block text-xs font-bold text-slate-800">{opt.name}</span>
										<span class="block text-[10px] text-slate-400 font-semibold mt-0.5"
											>{opt.desc}</span
										>
										<span class="block text-[10px] text-[#881B1B] font-bold mt-1">{opt.count}</span>
									</div>
								</button>
							{/each}
						</div>
					</div>
				</div>

				<!-- Institutional Overview + Report Activity Log -->
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
					<!-- Institutional Overview -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div class="p-5 border-b border-slate-100 bg-slate-50/20 select-none">
							<h3 class="text-sm font-bold font-serif text-slate-905">Institutional Overview</h3>
							<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
								Academic year 2025–26 at a glance
							</p>
						</div>
						<div class="p-5 space-y-6">
							<!-- Stat cards -->
							<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 select-none">
								{#each institutionalStats as stat}
									<div
										class="text-center rounded-xl bg-slate-50 border border-slate-150 py-3.5 px-2"
									>
										<div class="text-xl font-bold font-serif {stat.color}">{stat.value}</div>
										<div
											class="text-[9px] font-bold text-slate-400 uppercase tracking-wider mt-1 leading-tight"
										>
											{stat.label}
										</div>
									</div>
								{/each}
							</div>

							<!-- Avg score by department (bar chart) -->
							<div>
								<span class="text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
									>Avg Credits by Department</span
								>
								<div class="mt-3 flex gap-2">
									<div
										class="flex flex-col justify-between h-36 text-[9px] font-bold text-slate-300 text-right shrink-0"
									>
										<span>100</span><span>75</span><span>50</span><span>25</span><span>0</span>
									</div>
									<div class="flex-1 min-w-0">
										<div class="relative h-36">
											<div class="absolute inset-0 flex flex-col justify-between">
												{#each ['a', 'b', 'c', 'd', 'e'] as line (line)}
													<div class="border-t border-dashed border-slate-100"></div>
												{/each}
											</div>
											<div class="relative flex items-end justify-between gap-2 h-full">
												{#each deptScores as d}
													<div class="flex-1 flex items-end justify-center h-full">
														<div
															class="w-full max-w-[26px] rounded-t bg-gradient-to-t from-[#881B1B] to-[#6b1414] transition-all"
															style="height: {Math.min(d.score, 100)}%"
															title="{d.dept}: {d.score} avg credits"
														></div>
													</div>
												{/each}
											</div>
										</div>
										<div class="flex justify-between gap-2 mt-1.5">
											{#each deptScores as d}
												<span class="flex-1 text-center text-[9px] font-bold text-slate-400"
													>{d.dept}</span
												>
											{/each}
										</div>
									</div>
								</div>
							</div>

							<!-- Monthly participation trend (dual line) -->
							<div>
								<div class="flex items-center justify-between">
									<span class="text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
										>Monthly Participation Trend</span
									>
									<div class="flex items-center gap-3">
										<span class="flex items-center gap-1 text-[9px] font-bold text-slate-500">
											<span class="w-2.5 h-0.5 rounded-full bg-[#881B1B]"></span>Activities
										</span>
										<span class="flex items-center gap-1 text-[9px] font-bold text-slate-500">
											<span class="w-2.5 h-0.5 rounded-full bg-emerald-500"></span>Certificates
										</span>
									</div>
								</div>
								<div class="mt-3 flex gap-2">
									<div
										class="flex flex-col justify-between h-24 text-[9px] font-bold text-slate-300 text-right shrink-0"
									>
										{#each trendAxisLabels as lbl, lblIndex (lblIndex)}
											<span>{lbl}</span>
										{/each}
									</div>
									<div class="flex-1 min-w-0">
										<div class="relative h-24">
											<div class="absolute inset-0 flex flex-col justify-between">
												{#each ['a', 'b', 'c', 'd'] as line (line)}
													<div class="border-t border-dashed border-slate-100"></div>
												{/each}
											</div>
											<svg
												viewBox="0 0 100 40"
												preserveAspectRatio="none"
												class="relative w-full h-full overflow-visible"
											>
												<polyline
													points={activitiesPoints}
													fill="none"
													stroke="#881B1B"
													stroke-width="1.5"
													stroke-linecap="round"
													stroke-linejoin="round"
													vector-effect="non-scaling-stroke"
												/>
												<polyline
													points={certificatesPoints}
													fill="none"
													stroke="#10b981"
													stroke-width="1.5"
													stroke-linecap="round"
													stroke-linejoin="round"
													vector-effect="non-scaling-stroke"
												/>
											</svg>
										</div>
										<div class="flex justify-between gap-2 mt-1.5">
											{#each participationMonths as m (m)}
												<span class="flex-1 text-center text-[9px] font-bold text-slate-400"
													>{m}</span
												>
											{/each}
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Report Activity Log -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div
							class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/20 select-none"
						>
							<div>
								<h3 class="text-sm font-bold font-serif text-slate-905">Report Activity Log</h3>
								<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
									Audit trail of recent report actions
								</p>
							</div>
							<button
								type="button"
								onclick={() => triggerToast('Loading full activity log...')}
								class="px-3 py-1.5 border border-slate-200 hover:bg-slate-50 text-slate-600 font-bold text-[10px] uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
							>
								View All
							</button>
						</div>
						<div class="overflow-x-auto">
							<table class="w-full text-left border-collapse">
								<thead>
									<tr
										class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
									>
										<th class="py-3 px-5">Action</th>
										<th class="py-3 px-4">User</th>
										<th class="py-3 px-5">Timestamp</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-100 text-xs font-sans">
									{#each reportAuditLog as entry, entryIndex (entry.action + entryIndex)}
										<tr class="hover:bg-slate-50/30 transition-colors">
											<td class="py-3.5 px-5">
												<div class="flex items-center gap-2.5">
													<span class="w-4 h-4 rounded-full border-2 border-slate-200 shrink-0"
													></span>
													<span class="font-bold text-slate-800">{entry.action}</span>
												</div>
											</td>
											<td class="py-3.5 px-4 text-slate-600 font-semibold whitespace-nowrap"
												>{entry.user}</td
											>
											<td class="py-3.5 px-5 text-slate-500 font-semibold whitespace-nowrap"
												>{entry.time}</td
											>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			{:else if currentTab === 'System Settings'}
				<!-- System Settings (Step 7) -->
				<!-- Overview Stat Cards -->
				<section
					class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 select-none"
					aria-label="System settings overview"
				>
					<!-- Academic Year -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">2025–26</span>
							<div class="p-2.5 rounded-lg bg-blue-50 text-blue-600 border border-blue-100">
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
										d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Academic Year
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Current cycle
							</p>
						</div>
					</div>

					<!-- Credit Policies -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900"
								>{(settingsSections['Credit Policy']?.length ?? 0) + 3}</span
							>
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
										d="M16.5 18.75h-9m9 0a3 3 0 0 1 3 3h-15a3 3 0 0 1 3-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497m5.007 0a7.454 7.454 0 0 1-.982-3.172M9.497 14.25a7.454 7.454 0 0 0 .981-3.172M5.25 4.236c-.982.143-1.954.317-2.916.52A6.003 6.003 0 0 0 7.73 9.728M5.25 4.236V4.5c0 2.108.966 3.99 2.48 5.228M5.25 4.236V2.721C7.456 2.41 9.71 2.25 12 2.25c2.291 0 4.545.16 6.75.47v1.516M7.73 9.728a6.726 6.726 0 0 0 2.748 1.35m8.272-6.842V4.5c0 2.108-.966 3.99-2.48 5.228m2.48-5.492a46.32 46.32 0 0 1 2.916.52 6.003 6.003 0 0 1-5.395 4.972m0 0a6.726 6.726 0 0 1-2.749 1.35m0 0a6.772 6.772 0 0 1-3.044 0"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Credit Policies
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Rules active
							</p>
						</div>
					</div>

					<!-- Required Activities -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-slate-900">12</span>
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
										d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								Required Activities
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								Configured
							</p>
						</div>
					</div>

					<!-- System Status -->
					<div
						class="bg-white p-5 rounded-xl border border-slate-200 flex flex-col justify-between shadow-xs hover:shadow-md transition-shadow duration-200"
					>
						<div class="flex items-center justify-between">
							<span class="text-2xl font-bold font-serif text-emerald-600">Active</span>
							<div
								class="p-2.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100"
							>
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
										d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z"
									/>
								</svg>
							</div>
						</div>
						<div class="mt-4">
							<h3 class="text-xs font-bold text-slate-800 tracking-wide font-sans">
								System Status
							</h3>
							<p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-wider">
								All systems normal
							</p>
						</div>
					</div>
				</section>

				<!-- Overview + Platform Summary grid -->
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
					<!-- System Settings Overview (col-span-2) -->
					<div
						class="lg:col-span-2 bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden flex flex-col"
					>
						<!-- Header -->
						<div
							class="p-5 border-b border-slate-100 flex items-center justify-between gap-3 bg-slate-50/20 select-none"
						>
							<div>
								<h3 class="text-sm font-bold font-serif text-slate-905">
									System Settings Overview
								</h3>
								<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
									Manage platform-wide configuration settings.
								</p>
							</div>
							<button
								type="button"
								onclick={handleSaveAllSettings}
								class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none shrink-0"
							>
								Save Changes
							</button>
						</div>

						<!-- Section Tabs -->
						<div
							class="px-5 pt-3 border-b border-slate-100 flex flex-wrap gap-1.5 bg-white select-none"
						>
							{#each settingsTabs as tab}
								<button
									type="button"
									onclick={() => (activeSettingsTab = tab)}
									class="px-3.5 py-2 rounded-t-lg text-xs font-bold transition-all border-b-2 -mb-px
									{activeSettingsTab === tab
										? 'text-[#881B1B] border-[#881B1B]'
										: 'text-slate-500 border-transparent hover:text-slate-800'}"
								>
									{tab}
								</button>
							{/each}
						</div>

						<!-- Settings Table -->
						<div class="overflow-x-auto flex-grow">
							<table class="w-full text-left border-collapse">
								<thead>
									<tr
										class="border-b border-slate-150 bg-slate-50/50 text-[10px] font-extrabold text-slate-405 uppercase tracking-wider"
									>
										<th class="py-3.5 px-5">Setting Name</th>
										<th class="py-3.5 px-5">Description</th>
										<th class="py-3.5 px-5">Value</th>
										<th class="py-3.5 px-5">Status</th>
										<th class="py-3.5 px-5 text-center">Actions</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-slate-100 text-xs font-sans">
									{#each currentSettings as setting, index (setting.name)}
										<tr class="hover:bg-slate-50/30 transition-colors">
											<td class="py-4 px-5 font-bold text-slate-800 align-top">{setting.name}</td>
											<td class="py-4 px-5 text-slate-500 font-semibold align-top max-w-xs">
												{setting.description}
											</td>
											<td class="py-4 px-5 font-bold text-slate-800 align-top whitespace-nowrap">
												{setting.value}
											</td>
											<td class="py-4 px-5 align-top">
												<span
													class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border {settingStatusClass(
														setting.status
													)}"
												>
													<span class="w-1.5 h-1.5 rounded-full {settingStatusDot(setting.status)}"
													></span>
													{setting.status}
												</span>
											</td>
											<td class="py-4 px-5 align-top">
												<div class="flex items-center justify-center">
													<button
														type="button"
														onclick={() => openEditSetting(index)}
														aria-label="Edit {setting.name}"
														class="p-1.5 border border-slate-200 text-slate-500 hover:text-[#881B1B] hover:bg-slate-100 rounded-lg transition-colors focus:outline-none"
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
																d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Z"
															/>
														</svg>
													</button>
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>

						<!-- Footer -->
						<div
							class="p-4 border-t border-slate-100 bg-slate-50/30 flex items-center justify-between gap-3 select-none"
						>
							<span class="text-slate-500 font-semibold text-[11px]">
								{currentSettings.length} settings in this section
							</span>
							<button
								type="button"
								onclick={handleResetSettings}
								class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-600 font-bold text-[11px] uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
							>
								Reset to Default
							</button>
						</div>
					</div>

					<!-- Platform Summary (col-span-1) -->
					<div class="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden">
						<div class="p-5 border-b border-slate-100 bg-slate-50/20 select-none">
							<h3 class="text-sm font-bold font-serif text-slate-905">Platform Summary</h3>
							<p class="text-[11px] text-slate-500 font-semibold mt-0.5">
								Global configuration overview
							</p>
						</div>
						<div class="p-5 space-y-1">
							<div
								class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0"
							>
								<span class="text-xs font-semibold text-slate-500">Academic Year</span>
								<span class="text-xs font-bold text-slate-900">2025–26</span>
							</div>
							<div
								class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0"
							>
								<span class="text-xs font-semibold text-slate-500">Credit Policies</span>
								<span class="text-xs font-bold text-slate-900"
									>{(settingsSections['Credit Policy']?.length ?? 0) + 3}</span
								>
							</div>
							<div
								class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0"
							>
								<span class="text-xs font-semibold text-slate-500">Activity Rules</span>
								<span class="text-xs font-bold text-slate-900">12</span>
							</div>
							<div
								class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0"
							>
								<span class="text-xs font-semibold text-slate-500">Notif. Templates</span>
								<span class="text-xs font-bold text-slate-900">5</span>
							</div>
							<div class="flex items-center justify-between py-2.5">
								<span class="text-xs font-semibold text-slate-500">System Status</span>
								<span class="text-xs font-bold text-emerald-600">Active</span>
							</div>

							<button
								type="button"
								onclick={() => triggerToast('Opening global settings editor...')}
								class="w-full mt-3 py-2.5 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors focus:outline-none"
							>
								Edit Global Settings
							</button>
						</div>
					</div>
				</div>
			{:else}
				<!-- Under Construction placeholder for other tabs -->
				<div
					class="bg-white border border-slate-200 rounded-xl p-8 sm:p-12 text-center max-w-2xl mx-auto my-12 shadow-xs space-y-6"
				>
					<div
						class="w-16 h-16 bg-[#881B1B]/10 text-[#881B1B] flex items-center justify-center rounded-full mx-auto"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="2"
							stroke="currentColor"
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
						<h2 class="text-xl font-bold font-serif text-[#0b1535]">
							{currentTab} Under Construction
						</h2>
						<p class="text-xs text-slate-500 max-w-md mx-auto leading-relaxed">
							The features for the <strong>{currentTab}</strong> module are scheduled for development
							by other contributors.
						</p>
					</div>
					<button
						onclick={() => (currentTab = 'Dashboard')}
						class="px-5 py-2.5 bg-[#881B1B] hover:bg-[#881B1B]/95 text-white font-bold text-xs uppercase tracking-wider rounded-lg transition-colors"
					>
						Back to Dashboard
					</button>
				</div>
			{/if}
		</main>

		<!-- ==================== TOAST NOTIFICATION CONTAINER (Step 6) ==================== -->
		<div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2 max-w-sm">
			{#each toasts as toast (toast.id)}
				<div
					transition:slide={{ duration: 150 }}
					class="p-4 bg-slate-800 border border-slate-700 text-white rounded-xl shadow-2xl flex items-center justify-between gap-3 text-xs font-semibold font-sans min-w-[280px]"
				>
					<div class="flex items-center gap-2">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="2"
							stroke="currentColor"
							class="w-4 h-4 text-emerald-400 shrink-0"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M9 12.75 11.25 15 15 9.75M21 12c0 1.268-.63 2.39-1.593 3.068a3.745 3.745 0 01-1.043 3.296 3.745 3.745 0 01-3.296 1.043A3.745 3.745 0 0112 21c-1.268 0-2.39-.63-3.068-1.593a3.746 3.746 0 01-3.296-1.043 3.745 3.745 0 01-1.043-3.296A3.745 3.745 0 013 12c0-1.268.63-2.39 1.593-3.068a3.745 3.745 0 011.043-3.296 3.746 3.746 0 013.296-1.043A3.746 3.746 0 0112 3c1.268 0 2.39.63 3.068 1.593a3.746 3.746 0 013.296 1.043 3.746 3.746 0 011.043 3.296A3.745 3.745 0 0121 12Z"
							/>
						</svg>
						<span>{toast.message}</span>
					</div>
					<button
						type="button"
						onclick={() => dismissToast(toast.id)}
						class="text-slate-400 hover:text-white transition-colors focus:outline-none p-0.5 rounded-lg hover:bg-slate-700/50 shrink-0"
						aria-label="Dismiss notification"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="2.5"
							stroke="currentColor"
							class="w-3.5 h-3.5"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			{/each}
		</div>

		<!-- ==================== MODALS LIST (Step 6) ==================== -->
		<!-- Create User Modal -->
		{#if isCreateUserModalOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isCreateUserModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<form
					onsubmit={handleCreateUser}
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">Create New User</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								Register Student or Admin
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isCreateUserModalOpen = false)}
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
								for="new-username"
								class="text-[10px] font-extrabold text-slate-650 tracking-wider">FULL NAME *</label
							>
							<input
								id="new-username"
								type="text"
								bind:value={newUserName}
								placeholder="e.g. Amit Kumar"
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div class="flex flex-col gap-1.5">
								<label
									for="new-user-role"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider">ROLE *</label
								>
								<select
									id="new-user-role"
									bind:value={newUserRole}
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
								>
									<option value="Student">Student</option>
									<option value="Admin">Admin</option>
								</select>
							</div>

							<div class="flex flex-col gap-1.5">
								<label
									for="new-user-dept"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider">DEPARTMENT</label
								>
								<input
									id="new-user-dept"
									type="text"
									bind:value={newUserDept}
									placeholder="e.g. Computer Science"
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div class="flex flex-col gap-1.5">
								<label
									for="new-user-id"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider"
									>{newUserRole === 'Student' ? 'ROLL NUMBER *' : 'ADMIN ID *'}</label
								>
								<input
									id="new-user-id"
									type="text"
									bind:value={newUserId}
									placeholder={newUserRole === 'Student' ? 'e.g. IT2K24012' : 'e.g. admin.cs'}
									required
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>

							<div class="flex flex-col gap-1.5">
								<label
									for="new-user-email"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider">EMAIL *</label
								>
								<input
									id="new-user-email"
									type="email"
									bind:value={newUserEmail}
									placeholder="e.g. amit@iips.edu"
									required
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>
						</div>

						{#if newUserRole === 'Student'}
							<div class="flex flex-col gap-1.5">
								<label
									for="new-user-semester"
									class="text-[10px] font-extrabold text-slate-650 tracking-wider">SEMESTER</label
								>
								<input
									id="new-user-semester"
									type="number"
									min="1"
									max="10"
									bind:value={newUserSemester}
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>
						{/if}

						{#if createUserError}
							<div
								class="p-3 bg-red-50 border border-red-200 rounded-lg text-[11px] font-semibold text-red-700"
								role="alert"
							>
								{createUserError}
							</div>
						{/if}
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isCreateUserModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={creatingUser}
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{creatingUser ? 'Creating…' : 'Create User'}
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- View User Details Modal -->
		{#if isViewUserModalOpen && viewingUser}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isViewUserModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<div
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">User Profile Details</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								ID: {viewingUser.id}
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isViewUserModalOpen = false)}
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

					<div class="p-6 space-y-1">
						<div class="flex items-center justify-between py-2.5 border-b border-slate-100">
							<span class="text-xs font-semibold text-slate-500">Full Name</span>
							<span class="text-xs font-bold text-slate-900">{viewingUser.name}</span>
						</div>
						<div class="flex items-center justify-between py-2.5 border-b border-slate-100">
							<span class="text-xs font-semibold text-slate-500">Role</span>
							<span class="text-xs font-bold text-slate-900"
								>{viewingUser.displayRole || viewingUser.role}</span
							>
						</div>
						{#if viewingUser.email}
							<div class="flex items-center justify-between py-2.5 border-b border-slate-100">
								<span class="text-xs font-semibold text-slate-500">Email Address</span>
								<span class="text-xs font-bold text-slate-900 select-all">{viewingUser.email}</span>
							</div>
						{/if}
						<div class="flex items-center justify-between py-2.5 border-b border-slate-100">
							<span class="text-xs font-semibold text-slate-500">Department / Batch</span>
							<span class="text-xs font-bold text-slate-900">{viewingUser.dept}</span>
						</div>
						{#if viewingUser.role === 'Student'}
							<div class="flex items-center justify-between py-2.5 border-b border-slate-100">
								<span class="text-xs font-semibold text-slate-500">Current Semester</span>
								<span class="text-xs font-bold text-slate-900">{viewingUser.semester || 1}</span>
							</div>
						{/if}
						<div class="flex items-center justify-between py-2.5">
							<span class="text-xs font-semibold text-slate-500">Account Status</span>
							<span
								class="inline-flex items-center gap-1.5 font-bold {viewingUser.status === 'Active'
									? 'text-emerald-600'
									: 'text-slate-400'}"
							>
								<span
									class="w-1.5 h-1.5 rounded-full shrink-0 {viewingUser.status === 'Active'
										? 'bg-emerald-600'
										: 'bg-slate-400'}"
								></span>
								{viewingUser.status}
							</span>
						</div>
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isViewUserModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Close
						</button>
						<button
							type="button"
							onclick={() => {
								if (viewingUser) {
									isViewUserModalOpen = false;
									openEditUser(viewingUser);
								}
							}}
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Edit Profile
						</button>
					</div>
				</div>
			</div>
		{/if}

		<!-- Edit User Modal -->
		{#if isEditUserModalOpen && editingUser}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isEditUserModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<form
					onsubmit={handleSaveEditUser}
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">Edit User Profile</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								ID: {editingUser.id}
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isEditUserModalOpen = false)}
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

					<div class="p-6 space-y-4 font-sans">
						<div class="flex flex-col gap-1.5">
							<label
								for="edit-name"
								class="text-[10px] font-extrabold text-slate-650 tracking-wider">FULL NAME *</label
							>
							<input
								id="edit-name"
								type="text"
								bind:value={editUserName}
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>

						<div class="flex flex-col gap-1.5">
							<label
								for="edit-email"
								class="text-[10px] font-extrabold text-slate-650 tracking-wider"
								>EMAIL ADDRESS *</label
							>
							<input
								id="edit-email"
								type="email"
								bind:value={editUserEmail}
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>

						<div class="flex flex-col gap-1.5">
							<label
								for="edit-dept"
								class="text-[10px] font-extrabold text-slate-650 tracking-wider"
								>DEPARTMENT / BATCH *</label
							>
							<input
								id="edit-dept"
								type="text"
								bind:value={editUserDept}
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>

						{#if editingUser.role === 'Student'}
							<div class="grid grid-cols-2 gap-4">
								<div class="flex flex-col gap-1.5">
									<label
										for="edit-semester"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider"
										>SEMESTER *</label
									>
									<select
										id="edit-semester"
										bind:value={editUserSemester}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
									>
										{#each Array.from({ length: 12 }, (_, i) => i + 1) as sem}
											<option value={sem}>Semester {sem}</option>
										{/each}
									</select>
								</div>

								<div class="flex flex-col gap-1.5">
									<label
										for="edit-status"
										class="text-[10px] font-extrabold text-slate-650 tracking-wider">STATUS</label
									>
									<select
										id="edit-status"
										bind:value={editUserStatus}
										class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-850 bg-white focus:outline-none focus:border-slate-355"
									>
										<option value="Active">Active</option>
										<option value="Inactive">Inactive</option>
										<option value="Pending">Pending</option>
									</select>
								</div>
							</div>
						{/if}

						{#if editUserError}
							<div
								class="p-3 bg-red-50 text-red-700 text-xs font-semibold rounded-lg border border-red-100"
							>
								{editUserError}
							</div>
						{/if}
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isEditUserModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={editingUserSubmit}
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none disabled:opacity-50"
						>
							{editingUserSubmit ? 'Saving…' : 'Save Changes'}
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Create Activity Modal -->
		{#if isCreateActivityModalOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isCreateActivityModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<form
					onsubmit={handleCreateActivity}
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">Publish New Activity</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								Register Extracurricular Activity
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isCreateActivityModalOpen = false)}
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
								for="new-act-name"
								class="text-[10px] font-extrabold text-slate-655 tracking-wider"
								>ACTIVITY NAME *</label
							>
							<input
								id="new-act-name"
								type="text"
								bind:value={newActivityName}
								placeholder="e.g. Cyber Security Summit"
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>

						<div class="flex flex-col gap-1.5">
							<label
								for="new-act-credits"
								class="text-[10px] font-extrabold text-slate-655 tracking-wider"
								>CREDITS OFFERED *</label
							>
							<input
								id="new-act-credits"
								type="number"
								bind:value={newActivityCredits}
								min="1"
								max="50"
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isCreateActivityModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Publish Activity
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Create Track Modal -->
		{#if isCreateTrackModalOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isCreateTrackModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<form
					onsubmit={handleCreateTrack}
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">Create New Track</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								Add specialized extracurricular track
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isCreateTrackModalOpen = false)}
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
								class="text-[10px] font-extrabold text-slate-655 tracking-wider"
								>TRACK TITLE *</label
							>
							<input
								id="new-track-name"
								type="text"
								bind:value={newTrackName}
								placeholder="e.g. Social Entrepreneurship & Innovations"
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isCreateTrackModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Create Track
						</button>
					</div>
				</form>
			</div>
		{/if}

		<!-- Generate Report Modal -->
		{#if isGenerateReportModalOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isGenerateReportModalOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<div
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-905">Generate System Report</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								Export extracurricular summaries
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isGenerateReportModalOpen = false)}
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
								for="report-type"
								class="text-[10px] font-extrabold text-slate-655 tracking-wider">REPORT TYPE</label
							>
							<select
								id="report-type"
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 bg-white focus:outline-none focus:border-slate-355"
							>
								<option value="activities">Student Activities Summary</option>
								<option value="credits">Credit Distribution Ledger</option>
								<option value="audits">Administrative Audit Trails</option>
							</select>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div class="flex flex-col gap-1.5">
								<label
									for="report-from"
									class="text-[10px] font-extrabold text-slate-655 tracking-wider">START DATE</label
								>
								<input
									id="report-from"
									type="date"
									value="2026-06-01"
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>
							<div class="flex flex-col gap-1.5">
								<label
									for="report-to"
									class="text-[10px] font-extrabold text-slate-655 tracking-wider">END DATE</label
								>
								<input
									id="report-to"
									type="date"
									value="2026-06-30"
									class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
								/>
							</div>
						</div>
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isGenerateReportModalOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="button"
							onclick={handleGenerateReport}
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Export Data
						</button>
					</div>
				</div>
			</div>
		{/if}

		<!-- Edit Setting Modal (Step 7) -->
		{#if isEditSettingOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				onclick={(e) => {
					if (e.target === e.currentTarget) isEditSettingOpen = false;
				}}
				transition:fade={{ duration: 150 }}
				class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
			>
				<form
					onsubmit={handleSaveSetting}
					class="w-full max-w-md bg-white border border-slate-200 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
				>
					<div
						class="p-5 border-b border-slate-150 flex items-center justify-between bg-slate-50/30"
					>
						<div>
							<h3 class="text-sm font-bold font-serif text-slate-900">Edit Setting</h3>
							<p class="text-[9px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
								{activeSettingsTab}
							</p>
						</div>
						<button
							type="button"
							onclick={() => (isEditSettingOpen = false)}
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
							<span class="text-[10px] font-extrabold text-slate-650 tracking-wider"
								>SETTING NAME</span
							>
							<span class="text-sm font-bold text-slate-800">{editSettingName}</span>
						</div>
						<div class="flex flex-col gap-1.5">
							<label
								for="edit-setting-value"
								class="text-[10px] font-extrabold text-slate-650 tracking-wider">VALUE *</label
							>
							<input
								id="edit-setting-value"
								type="text"
								bind:value={editSettingValue}
								required
								class="px-3 py-2 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-none focus:border-slate-355"
							/>
						</div>
					</div>

					<div
						class="p-5 border-t border-slate-150 flex items-center justify-end gap-2.5 bg-slate-50/30"
					>
						<button
							type="button"
							onclick={() => (isEditSettingOpen = false)}
							class="px-4 py-2 border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 bg-[#881B1B] hover:bg-[#881B1B]/90 text-white font-bold text-xs uppercase rounded-lg transition-colors focus:outline-none"
						>
							Save Setting
						</button>
					</div>
				</form>
			</div>
		{/if}
	</div>
</div>
