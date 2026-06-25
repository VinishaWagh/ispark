<script lang="ts">
	// Props configuration
	let {
		badgeText = 'Departmental Affiliated Cell',
		titleText = 'Skill, Personality Advancement & Refinement Cell',
		subtitleText = '',
		descriptionText = 'Welcome to ISPARC, an official cell of the Institute of Infrastructure and Professional Studies (IIPS), DAVV. We facilitate structured extracurricular activities, personality enhancement tracks, and grading systems to foster student growth and leadership.',
		primaryCtaText = 'Grading Matrix',
		primaryCtaHref = '#grading-scheme',
		secondaryCtaText = 'Explore Tracks',
		secondaryCtaHref = '#tracks'
	} = $props();

	// Top performers data
	const topPerformers = [
		{ name: 'Priya Sharma', points: 95, track: 'Track A' },
		{ name: 'Rahul Verma', points: 88, track: 'Track B' },
		{ name: 'Sneha Patel', points: 82, track: 'Track A' },
		{ name: 'Amit Kumar', points: 78, track: 'Track B' }
	];
</script>

<section class="relative py-16 sm:py-24 bg-white">
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-start">
			
			<!-- Left Column: Headline and Context -->
			<div class="lg:col-span-7 space-y-6">
				
				<!-- Academic Affiliation Badge -->
				<div class="inline-flex items-center px-3 py-1 bg-[#284B63] rounded-sm text-[10px] font-bold tracking-widest uppercase text-white">
					{badgeText}
				</div>

				<!-- Main Institutional Headline -->
				<h1 class="text-4xl sm:text-5xl lg:text-6xl font-bold text-[#881B1B] font-serif tracking-tight leading-none">
	{titleText}
</h1>

				<!-- Subtitle -->
				<h2 class="text-lg sm:text-xl font-semibold text-[#1e3a5f] font-sans tracking-tight">
					{subtitleText}
				</h2>

				<!-- Subtext explanations -->
				<p class="text-sm text-slate-600 leading-relaxed font-sans font-normal max-w-2xl">
					{descriptionText}
				</p>

				<!-- Buttons -->
				<div class="flex flex-col sm:flex-row gap-3 pt-2">
					<a
						href={primaryCtaHref}
						class="inline-flex items-center justify-center border-2 border-[#284B63] text-[#284B63] font-semibold text-xs tracking-wider uppercase px-6 py-3 hover:bg-[#284B63] hover:text-white transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#284B63] focus-visible:ring-offset-2"
					>
						{primaryCtaText}
					</a>
					<a
						href={secondaryCtaHref}
						class="inline-flex items-center justify-center border-2 border-[#284B63] text-[#284B63] font-semibold text-xs tracking-wider uppercase px-6 py-3 hover:bg-[#284B63] hover:text-white transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#284B63] focus-visible:ring-offset-2"
					>
						{secondaryCtaText}
					</a>
				</div>

			</div>

			<!-- Right Column: Top Performers -->
			<div class="lg:col-span-5 bg-white border border-slate-200 p-6 rounded-sm">
				<h2 class="text-lg font-bold font-serif text-[#1e3a5f] mb-4">
					Top Performers
				</h2>

				<div class="space-y-3">
					{#each topPerformers as performer, i}
						<div class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-b-0">
							<div class="flex items-center gap-3">
								<span class="w-6 h-6 rounded-full bg-[#284B63] text-white text-xs font-bold flex items-center justify-center">
									{i + 1}
								</span>
								<div>
									<span class="text-sm font-medium text-slate-800 font-sans block">{performer.name}</span>
									<span class="text-xs text-slate-500 font-sans">{performer.track}</span>
								</div>
							</div>
							<span class="text-sm font-bold text-[#284B63] font-sans">{performer.points} pts</span>
						</div>
					{/each}
				</div>
			</div>

		</div>
	</div>
</section>
