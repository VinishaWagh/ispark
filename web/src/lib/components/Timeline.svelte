<script lang="ts">
	// Timeline steps
	const steps = [
		{
			phase: '01',
			title: 'Awareness',
			desc: 'Department briefings and orientation sessions regarding the upcoming cell cycle.'
		},
		{
			phase: '02',
			title: 'Enrollment',
			desc: 'Registration of interest and selection of primary development tracks.'
		},
		{
			phase: '03',
			title: 'Workshops',
			desc: 'Direct instruction sessions in skill tracks and soft-skills seminars.'
		},
		{
			phase: '04',
			title: 'Project Application',
			desc: 'Implementation of learnings in verified, peer-evaluated projects.'
		},
		{
			phase: '05',
			title: 'Assessment',
			desc: 'Committee review of submitted logs, credentials, and achievements.'
		},
		{
			phase: '06',
			title: 'Certification',
			desc: 'Awarding of extra-curricular grades and university-backed certificates.'
		}
	];
</script>

<section class="py-16 sm:py-20 bg-bg-warm border-t border-border-gray">
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		
		<!-- Section Header -->
		<div class="pb-8 mb-12">
			<span class="text-[10px] font-bold tracking-[0.2em] text-accent-red uppercase font-sans">
				Process
			</span>
			<h2 class="text-2xl sm:text-3xl font-bold font-serif text-primary-dark tracking-tight mt-2">
				The Student Event Journey
			</h2>
			<p class="text-sm text-secondary-dark font-sans max-w-2xl mt-3 leading-relaxed">
				We enforce a structured advancement pipeline to ensure quality and accountability. Below is the official progression timeline for cell members.
			</p>
		</div>

		<!-- Timeline Container -->
		<div class="relative">
			<!-- Timeline path line (Desktop) -->
			<div class="hidden lg:block absolute top-[22px] left-0 right-0 h-0.5 bg-border-gray -z-10"></div>
			
			<!-- Timeline Grid -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-8 lg:gap-4">
				{#each steps as step, i}
					<div class="relative flex flex-col items-start gap-3">
						<!-- Timeline Connector Node -->
						<div class="flex items-center justify-between w-full lg:w-auto">
							<div class="w-11 h-11 bg-accent-red text-white font-bold text-sm flex items-center justify-center rounded-full shrink-0">
								{step.phase}
							</div>
						</div>

						<!-- Content -->
						<div class="space-y-1">
							<h3 class="text-sm font-bold text-primary-dark font-sans tracking-tight">
								{step.title}
							</h3>
							<p class="text-xs text-secondary-dark leading-relaxed font-sans">
								{step.desc}
							</p>
						</div>
					</div>
				{/each}
			</div>
		</div>

	</div>
</section>
