<script lang="ts">
	const rules = [
		{
			title: 'National / State Representation',
			desc: 'Up to 20 points per event for representing DAVV in national debate matches or athletic tournaments.'
		},
		{
			title: 'College Hackathons & Project Meets',
			desc: '5 to 15 points depending on rank and tier of involvement.'
		},
		{
			title: 'Social Service & Volunteering (NSS/NGO)',
			desc: '1 point per 10 hours of verified field work, up to a maximum of 25 points per semester.'
		}
	];
</script>

<section id="grading-scheme" class="py-16 sm:py-20 bg-white border-t border-slate-200">
	<div class="max-w-7xl mx-auto px-6 lg:px-8 space-y-16">
		
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-16 items-start">
			
			<div class="space-y-6">
				<span class="text-[10px] font-bold tracking-[0.2em] text-[#284B63] uppercase font-sans">
					Grading Scheme
				</span>
				<h2 class="text-2xl sm:text-3xl font-bold font-serif text-[#1e3a5f] tracking-tight">
					Accreditation & Grading
				</h2>
				<p class="text-sm text-slate-600 leading-relaxed font-sans">
					All student certificates and service logs are audited by the departmental committee before final grades are submitted to the university.
				</p>
					
				<blockquote class="border-l-2 border-[#284B63] pl-5 py-1">
					<p class="text-sm italic text-slate-700 font-sans leading-relaxed">
						"Personality development is not merely an addition; it is an essential component of professional readiness."
					</p>
					<cite class="text-[10px] font-bold tracking-wider text-slate-400 uppercase font-sans mt-3 block">
						— DAVV BOARD OF STUDIES GUIDELINES
					</cite>
				</blockquote>
			</div>

			<div class="bg-[#F8F8F8] border border-slate-200 p-6 rounded-sm space-y-5">
				<h3 class="text-lg font-bold font-serif text-[#1e3a5f]">
					Point Allocation Rules
				</h3>
				
				<ul class="space-y-5">
					{#each rules as rule}
						<li class="flex items-start gap-3">
							<!-- Red Square Bullet (Figma spec) -->
							<span class="w-2.5 h-2.5 bg-acad-red mt-1 shrink-0"></span>
							<div>
								<h4 class="text-sm font-semibold text-slate-800 font-sans">{rule.title}:</h4>
								<p class="text-xs text-slate-500 leading-relaxed font-sans mt-1">{rule.desc}</p>
							</div>
						</li>
					{/each}
				</ul>
			</div>

		</div>

		<div class="bg-[#284B63] text-white p-10 sm:p-14 text-center rounded-sm relative overflow-hidden">
			<div class="absolute top-0 left-0 right-0 h-1 bg-[#C89B3C]"></div>
			
			<div class="max-w-2xl mx-auto space-y-5">
				<h2 class="text-2xl sm:text-3xl font-bold font-serif tracking-tight">
					Student Registry Portal
				</h2>
				<p class="text-sm text-white/80 leading-relaxed font-sans">
					Access your student profile, upload credentials, view points status, and verify your semester extra-curricular grades.
				</p>
				
				<div class="pt-2">
					<a
						href="/portal"
						class="inline-flex items-center justify-center bg-acad-red hover:bg-acad-red/90 text-white font-bold text-xs tracking-wider uppercase px-8 py-3.5 border border-acad-red transition-colors duration-250 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-gold focus-visible:ring-offset-2"
					>
						Enter Portal
					</a>
				</div>
			</div>
		</div>

	</div>
</section>
