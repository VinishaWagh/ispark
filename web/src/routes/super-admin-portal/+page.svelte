<script lang="ts">
	import DevCredentials from '$lib/DevCredentials.svelte';
	import { goto } from '$app/navigation';
	import { onDestroy } from 'svelte';
	import { fade, slide } from 'svelte/transition';

	import { API_BASE_URL } from '$lib/config';

	const formId = 'superadmin-login';

	// Reactive state using Svelte 5 runes
	let superAdminId = $state('');
	let password = $state('');
	let showPassword = $state(false);

	// Simulated states
	let submitting = $state(false);
	let loginSuccess = $state(false);
	let errorMsg = $state('');

	// Rate Limiting & Lockout states
	let failedAttempts = $state(0);
	let lockoutTimeLeft = $state(0);
	let lockoutInterval: ReturnType<typeof setInterval> | undefined;

	function startLockout() {
		lockoutTimeLeft = 60;
		errorMsg = 'Too many failed login attempts. Please try again in 60 seconds.';

		if (lockoutInterval) clearInterval(lockoutInterval);

		lockoutInterval = setInterval(() => {
			lockoutTimeLeft -= 1;
			if (lockoutTimeLeft <= 0) {
				clearInterval(lockoutInterval);
				failedAttempts = 0;
				errorMsg = '';
			} else {
				errorMsg = `Too many failed login attempts. Please try again in ${lockoutTimeLeft} seconds.`;
			}
		}, 1000);
	}

	onDestroy(() => {
		if (lockoutInterval) clearInterval(lockoutInterval);
	});

	// Form validity derived state
	let isFormValid = $derived(superAdminId.trim() !== '' && password.trim() !== '');

	const footerLines = [
		'International Institute of Professional Studies (IIPS)',
		'Devi Ahilya Vishwavidyalaya (DAVV), Indore'
	];

	const currentYear = new Date().getFullYear();

	const featureItems = [
		{
			title: 'Complete Platform Access',
			description: 'Manage all modules, users, admins and configurations.',
			iconPath: 'cog'
		},
		{
			title: 'Role & Permission Control',
			description: 'Assign and monitor access levels across the system.',
			iconPath: 'users'
		},
		{
			title: 'System Analytics',
			description: 'Track activities, reports and institutional insights.',
			iconPath: 'chart'
		},
		{
			title: 'Security Monitoring',
			description: 'Review logs and maintain platform security.',
			iconPath: 'shield'
		}
	];

	function routeAfterLogin() {
		loginSuccess = true;
		setTimeout(() => {
			goto('/super-admin-portal/dashboard');
		}, 1000);
	}

	// Submit Handler
	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!isFormValid || lockoutTimeLeft > 0) return;

		submitting = true;
		errorMsg = '';

		try {
			const response = await fetch(`${API_BASE_URL}/api/admin/auth/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ admin_id: superAdminId, password: password })
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || 'Invalid Super Admin ID or Password.');
			}

			if (data.admin?.role !== 'superadmin') {
				throw new Error('This account does not have super admin access.');
			}

			localStorage.setItem('superadmin_token', data.access_token);
			failedAttempts = 0;
			routeAfterLogin();
		} catch (err) {
			failedAttempts += 1;
			if (failedAttempts >= 5) {
				startLockout();
			} else {
				errorMsg =
					err instanceof Error
						? err.message
						: `Invalid Super Admin ID or Password. (Attempt ${failedAttempts} of 5)`;
			}
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Super Admin Control Portal | iSPARC IIPS DAVV</title>
	<meta
		name="description"
		content="Secure access for iSPARC super administrators to manage the complete platform, user roles, system settings, and institutional operations."
	/>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="min-h-screen bg-bg-main flex flex-col items-center w-full">
	<!-- ==================== HEADER ==================== -->
	<header class="w-full bg-white border-b border-border-base sticky top-0 z-50 shadow-xs">
		<div class="max-w-6xl mx-auto flex items-center justify-between px-6 py-4">
			<!-- Brand Logo -->
			<a
				href="/"
				class="flex items-center gap-2 group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inst-navy focus-visible:ring-offset-2 rounded-lg transition-all duration-200"
			>
				<div class="flex flex-col justify-center">
					<div
						class="flex items-baseline text-2xl font-bold tracking-tight text-slate-900 leading-none"
					>
						<span>i</span><span class="text-acad-red font-extrabold">SPARC</span>
					</div>
				</div>
			</a>

			<!-- Back to Home Link -->
			<a
				href="/"
				aria-label="Back To Home"
				class="inline-flex items-center gap-2 focus:outline-none focus-visible:ring-2 focus-visible:ring-inst-navy focus-visible:ring-offset-2 rounded-md px-3 py-1.5 text-slate-650 hover:text-inst-navy transition-all duration-200"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="2"
					stroke="currentColor"
					class="w-4 h-4"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
					/>
				</svg>
				<span class="text-[13px] font-semibold uppercase tracking-wider"> Back To Home </span>
			</a>
		</div>
	</header>

	<!-- ==================== MAIN ==================== -->
	<main class="flex-grow w-full flex justify-center items-start">
		<section
			class="max-w-6xl w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-10 px-4 sm:px-6 lg:px-8 py-12 items-start font-sans"
		>
			<!-- Left Column: Super Admin Portal Description & System Privileges -->
			<aside class="lg:col-span-5 flex flex-col gap-6 animate-fade-in">
				<div class="flex flex-col gap-4">
					<div>
						<h1
							class="text-4xl font-extrabold text-acad-red font-serif leading-tight tracking-tight"
						>
							Super Admin<br />
							<span class="text-inst-navy">Control Portal</span>
						</h1>
					</div>

					<p class="text-slate-500 text-sm leading-relaxed">
						Secure access for iSPARC super administrators to manage the complete platform, user
						roles, system settings, and institutional operations.
					</p>
				</div>

				<!-- System Privileges Details -->
				<div
					class="flex flex-col gap-4 bg-white p-6 rounded-xl border border-border-base shadow-xs"
				>
					<h3 class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-2">
						System Privileges
					</h3>

					<div class="flex flex-col gap-5">
						{#each featureItems as item}
							<div class="flex items-start gap-4">
								<div
									class="flex w-9 h-9 items-center justify-center shrink-0 bg-slate-50 rounded-lg border border-border-base text-inst-navy"
								>
									{#if item.iconPath === 'cog'}
										<!-- Cog SVG -->
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
												d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
											/>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
											/>
										</svg>
									{:else if item.iconPath === 'users'}
										<!-- Users SVG -->
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
												d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0118 18.72zm-12 0a6.062 6.062 0 0112 0v.318c0 .243-.016.481-.049.714A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.72zm4.5-9a3 3 0 11-6 0 3 3 0 016 0zM18 9.75a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
											/>
										</svg>
									{:else if item.iconPath === 'chart'}
										<!-- Chart SVG -->
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
												d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
											/>
										</svg>
									{:else}
										<!-- Shield SVG -->
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
												d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
											/>
										</svg>
									{/if}
								</div>

								<div class="flex flex-col gap-0.5">
									<h4 class="text-sm font-bold text-slate-800">{item.title}</h4>
									<p class="text-xs text-slate-500 leading-normal">{item.description}</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</aside>

			<!-- Right Column: Super Admin Login Card -->
			<div
				class="lg:col-span-7 bg-white rounded-2xl border border-border-base shadow-sm overflow-hidden animate-fade-up"
			>
				{#if loginSuccess}
					<!-- Success View -->
					<div
						class="p-8 sm:p-12 text-center flex flex-col items-center gap-6"
						in:fade={{ duration: 400 }}
					>
						<div
							class="w-16 h-16 bg-emerald-100 text-emerald-800 flex items-center justify-center rounded-full border-2 border-emerald-250 animate-bounce"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="3"
								stroke-linecap="round"
								stroke-linejoin="round"
								class="w-8 h-8"
							>
								<polyline points="20 6 9 17 4 12" />
							</svg>
						</div>

						<div>
							<h2 class="text-2xl font-bold text-slate-900 font-serif">
								Authenticated Successfully!
							</h2>
							<p class="text-slate-500 text-sm mt-2 max-w-md mx-auto">
								Welcome, Super Admin <span class="font-semibold text-slate-800">{superAdminId}</span
								>. You have logged into the iSPARC Master Administration Panel. Redirecting to super
								admin console...
							</p>
						</div>

						<div class="flex gap-3 w-full max-w-sm pt-4">
							<a
								href="/super-admin-portal/dashboard"
								class="flex-1 py-3 text-center bg-inst-navy hover:bg-inst-navy/90 text-white font-semibold text-xs tracking-wider uppercase rounded-xl transition duration-200"
							>
								Go to Console
							</a>
						</div>
					</div>
				{:else}
					<!-- Super Admin Login Form -->
					<div class="p-6 sm:p-8 border-b border-border-base bg-slate-50/50">
						<div class="text-[10px] font-bold tracking-widest text-slate-400 uppercase">
							SUPER ADMIN LOGIN
						</div>
						<h2 class="text-2xl font-bold text-inst-navy font-serif leading-tight mt-1">
							Welcome Super Admin
						</h2>
						<p class="text-slate-500 text-xs mt-1">
							Login to access the iSPARC Master Administration Panel
						</p>
					</div>

					<form onsubmit={handleSubmit} class="p-6 sm:p-8 flex flex-col gap-6">
						<!-- Super Admin ID Input -->
						<div class="flex flex-col gap-1.5">
							<label
								for="{formId}-super-admin-id"
								class="text-[11px] font-bold text-slate-700 tracking-wider"
							>
								SUPER ADMIN ID
							</label>
							<input
								id="{formId}-super-admin-id"
								type="text"
								bind:value={superAdminId}
								disabled={lockoutTimeLeft > 0}
								placeholder="Enter your super admin ID"
								autoComplete="username"
								class="w-full px-3.5 py-2.5 bg-white rounded-lg border border-border-base text-[13px] text-slate-800 placeholder:text-slate-400 focus:outline-none focus:border-inst-navy focus:ring-2 focus:ring-inst-navy/10 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
								required
							/>
						</div>

						<!-- Password Input -->
						<div class="flex flex-col gap-1.5">
							<label
								for="{formId}-password"
								class="text-[11px] font-bold text-slate-700 tracking-wider"
							>
								PASSWORD
							</label>
							<div class="relative flex items-center">
								<input
									id="{formId}-password"
									type={showPassword ? 'text' : 'password'}
									bind:value={password}
									disabled={lockoutTimeLeft > 0}
									placeholder="Enter your password"
									autoComplete="current-password"
									class="w-full pl-3.5 pr-12 py-2.5 bg-white rounded-lg border border-border-base text-[13px] text-slate-800 placeholder:text-slate-400 focus:outline-none focus:border-inst-navy focus:ring-2 focus:ring-inst-navy/10 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
									required
								/>
								<button
									type="button"
									onclick={() => (showPassword = !showPassword)}
									class="absolute right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-650"
									aria-label={showPassword ? 'Hide password' : 'Show password'}
								>
									{#if showPassword}
										<!-- Eye-off SVG -->
										<svg
											xmlns="http://www.w3.org/2000/svg"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
											stroke-linecap="round"
											stroke-linejoin="round"
											class="w-5 h-5"
										>
											<line x1="2" y1="2" x2="22" y2="22" /><path
												d="M17.547 17.547a8.553 8.553 0 0 1-5.547 1.953 8.8 8.8 0 0 1-8.547-5.5 10.87 10.87 0 0 1 1.761-3.239"
											/><path
												d="M9.88 4.22a8.8 8.8 0 0 1 1.62-.22 8.82 8.82 0 0 1 8.547 5.5 10.64 10.64 0 0 1-1.341 2.871"
											/><circle cx="12" cy="12" r="3" />
										</svg>
									{:else}
										<!-- Eye SVG -->
										<svg
											xmlns="http://www.w3.org/2000/svg"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
											stroke-linecap="round"
											stroke-linejoin="round"
											class="w-5 h-5"
										>
											<path
												d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0z"
											/><circle cx="12" cy="12" r="3" />
										</svg>
									{/if}
								</button>
							</div>
						</div>

						<!-- Forgot Password -->
						<div class="flex items-center justify-between text-xs pt-1">
							<div></div>

							<button
								type="button"
								onclick={() =>
									alert(
										'Please contact the system administrator to retrieve or reset your super admin credentials.'
									)}
								class="font-bold text-acad-red hover:underline transition-colors focus:outline-none"
							>
								Forgot Password?
							</button>
						</div>

						{#if errorMsg}
							<div
								class="p-3 bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg font-semibold animate-fade-in"
								transition:slide
							>
								{errorMsg}
							</div>
						{/if}

						<!-- Submit Button -->
						<button
							type="submit"
							disabled={submitting || !isFormValid || lockoutTimeLeft > 0}
							class="w-full py-3.5 bg-inst-navy hover:bg-inst-navy/90 text-white font-bold text-sm tracking-widest uppercase rounded-xl transition duration-200 shadow-sm disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center"
						>
							{#if submitting}
								<svg
									class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
									fill="none"
									viewBox="0 0 24 24"
								>
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
								LOGGING IN...
							{:else}
								LOGIN
							{/if}
						</button>
					</form>
					<DevCredentials portal="superadmin" />
				{/if}
			</div>
		</section>
	</main>

	<!-- ==================== FOOTER ==================== -->
	<footer class="w-full bg-inst-navy py-8 border-t border-slate-900 mt-auto">
		<div class="max-w-6xl mx-auto flex flex-col items-center gap-1.5 px-6 text-center">
			<h2 class="text-base font-bold text-white tracking-wider uppercase font-sans">iSPARC</h2>
			{#each footerLines as line}
				<p class="text-xs font-normal text-white/50 leading-relaxed font-sans">
					{line}
				</p>
			{/each}
			<p class="text-[10px] text-white/30 mt-2 font-sans">
				© {currentYear} IIPS DAVV. All rights reserved.
			</p>
		</div>
	</footer>
</div>

<style>
	.animate-fade-in {
		opacity: 0;
		transform: translateY(-10px);
		animation: fadeIn 0.8s ease-out forwards;
	}

	.animate-fade-up {
		opacity: 0;
		transform: translateY(25px);
		animation: fadeUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
		animation-delay: 0.15s;
	}

	@keyframes fadeIn {
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes fadeUp {
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
