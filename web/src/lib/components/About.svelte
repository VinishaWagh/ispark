<script lang="ts">
	// Pillars data
	const pillars = [
		{
			num: '01',
			title: 'Structured Skill Tracks',
			desc: 'Participate in verified self-learning paths. We cover Technical Web Systems, Public Speaking & Oratory, NSS Social Service, and Athletics.'
		},
		{
			num: '02',
			title: 'Scheduled Cell Activities',
			desc: 'Register for campus activities, including hackathons, debating forums, blood donation camps, and inter-department sports meets.'
		},
		{
			num: '03',
			title: 'Student Score Registry',
			desc: 'Submit participation certificates and verified logs directly to the cell board to update your extra-curricular tally and view academic leaderboards.'
		}
	];
</script>

<section id="activities" class="py-16 sm:py-20 bg-[#F8F8F8] border-t border-slate-200">
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		
		<!-- Section Header -->
		<div class="pb-8 mb-12">
			<span class="text-[10px] font-bold tracking-[0.2em] text-[#881B1B] uppercase font-sans">
				About
			</span>
			<h2 class="text-2xl sm:text-3xl font-bold font-serif text-primary-dark tracking-tight mt-2">
				Cell Foundation Pillars
			</h2>
			<p class="text-sm text-slate-500 font-sans max-w-2xl mt-3 leading-relaxed">
				Through peer learning, campus representation, and social responsibility, we guide students to construct robust portfolios.
			</p>
		</div>

		<!-- Pillars Grid (3 Columns) -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
			{#each pillars as pillar}
				<div class="bg-white p-6 rounded-sm border border-slate-200 space-y-4">
					<!-- Large Academic Index -->
					<div class="text-4xl font-bold text-[#C89B3C] font-sans tracking-wide">
						{pillar.num}
					</div>
					<!-- Title -->
					<h3 class="text-base font-bold text-primary-dark font-sans tracking-tight">
						{pillar.title}
					</h3>
					<!-- Description -->
					<p class="text-sm text-slate-600 leading-relaxed font-sans font-normal">
						{pillar.desc}
					</p>
				</div>
			{/each}
		</div>

	</div>
</section>
