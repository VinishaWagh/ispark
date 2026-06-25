<script lang="ts">
	import { page } from '$app/state';
	import { slide } from 'svelte/transition';

	// Define interface for navigation links
	interface NavItem {
		label: string;
		href: string;
	}

	// Props configuration using Svelte 5 runes
	let {
		navItems = [
			{ label: 'Overview', href: '/' },
			{ label: 'Skill Tracks', href: '#tracks' },
			{ label: 'Activities', href: '#activities' },
			{ label: 'Leaderboard', href: '#leaderboard' },
			{ label: 'Grading Scheme', href: '#grading-scheme' }
		] as NavItem[],
		loginHref = '/login'
	} = $props();

	// Reactive state for the mobile menu drawer
	let isOpen = $state(false);

	function toggleMenu() {
		isOpen = !isOpen;
	}

	function closeMenu() {
		isOpen = false;
	}

	// Helper to check if a navigation item is active
	function isActive(href: string): boolean {
		const path = page.url.pathname;
		if (href === '/') {
			return path === '/';
		}
		if (href.startsWith('#')) {
			// In production, we're on the same page, active state is handled by hashes or scroll
			return false;
		}
		return path === href || path.startsWith(href + '/');
	}
</script>

<!-- Sticky wrapper -->
<header class="sticky top-0 z-50 w-full bg-white/95 backdrop-blur-md border-b border-border-gray">
	<!-- Gold accent border on top -->
	<div class="h-[3px] bg-[#C89B3C]"></div>
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		<div class="flex items-center justify-between h-[72px]">
			
			<!-- Left Section: Academic Logo & Branding -->
			<a href="/" onclick={closeMenu} class="flex items-center gap-2.5 group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold focus-visible:ring-offset-2 rounded-lg transition-all duration-200">
				<div class="flex flex-col justify-center">
					<div class="flex items-baseline text-2xl font-bold tracking-tight text-primary-dark font-sans leading-none">
						<span class="text-slate-900">i</span><span class="text-acad-red">SPARC</span>
					</div>
				</div>
			</a>

			<!-- Center Section: Desktop Navigation Links (Uppercase, institutional) -->
			<nav class="hidden md:flex items-center gap-7" aria-label="Desktop Navigation">
				<ul class="flex items-center gap-7">
					{#each navItems as item}
						<li>
							<a
								href={item.href}
								class="text-xs font-bold tracking-wider uppercase transition-all duration-200 hover:text-accent-gold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold focus-visible:ring-offset-1 rounded-md px-2 py-1 font-sans {isActive(item.href) ? 'text-accent-gold border-b border-accent-gold' : 'text-slate-600'}"
								aria-current={isActive(item.href) ? 'page' : undefined}
							>
								{item.label}
							</a>
						</li>
					{/each}
				</ul>
			</nav>

			<!-- Right Section: Desktop Portal Login (Outlined academic style) -->
			<div class="hidden md:flex items-center">
				<a
					href={loginHref}
					class="inline-flex items-center justify-center border border-slate-300 hover:border-slate-800 text-[13px] font-semibold text-slate-800 px-6 py-2 rounded-md hover:bg-slate-900 hover:text-white transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold focus-visible:ring-offset-2 font-sans"
				>
					Portal Login
				</a>
			</div>

			<!-- Mobile Menu Button (Hamburger) -->
			<div class="flex md:hidden items-center">
				<button
					onclick={toggleMenu}
					type="button"
					class="relative flex flex-col justify-center items-center w-10 h-10 rounded-lg text-slate-800 hover:bg-slate-100 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold"
					aria-label={isOpen ? 'Close menu' : 'Open menu'}
					aria-expanded={isOpen}
					aria-controls="mobile-menu"
				>
					<span class="sr-only">Toggle Menu</span>
					<!-- Animated Hamburger Bars -->
					<span class="w-5.5 h-0.5 bg-current rounded-full transition-all duration-300 {isOpen ? 'rotate-45 translate-y-[6px]' : ''}"></span>
					<span class="w-5.5 h-0.5 bg-current rounded-full my-[5px] transition-all duration-300 {isOpen ? 'opacity-0' : ''}"></span>
					<span class="w-5.5 h-0.5 bg-current rounded-full transition-all duration-300 {isOpen ? '-rotate-45 -translate-y-[6px]' : ''}"></span>
				</button>
			</div>

		</div>
	</div>

	<!-- Mobile Dropdown Menu -->
	{#if isOpen}
		<div
			transition:slide={{ duration: 200 }}
			class="md:hidden border-t border-border-gray bg-bg-warm shadow-md overflow-hidden"
			id="mobile-menu"
		>
			<nav class="px-6 py-5 space-y-4" aria-label="Mobile Navigation">
				<ul class="space-y-1">
					{#each navItems as item}
						<li>
							<a
								href={item.href}
								onclick={closeMenu}
								class="block w-full px-3 py-2.5 rounded-md text-sm font-bold tracking-wider uppercase transition-all duration-200 font-sans {isActive(item.href) ? 'bg-slate-100 text-accent-gold' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
								aria-current={isActive(item.href) ? 'page' : undefined}
							>
								{item.label}
							</a>
						</li>
					{/each}
				</ul>
				
				<!-- Separator -->
				<div class="h-px bg-slate-200 my-4"></div>

				<!-- Mobile CTAs -->
				<div class="flex flex-col gap-2.5 px-3 pb-2">
					<a
						href={loginHref}
						onclick={closeMenu}
						class="flex items-center justify-center w-full py-2.5 rounded-md text-sm font-semibold border border-slate-350 text-slate-800 hover:bg-slate-900 hover:text-white transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold font-sans"
					>
						Portal Login
					</a>
				</div>
			</nav>
		</div>
	{/if}
</header>
