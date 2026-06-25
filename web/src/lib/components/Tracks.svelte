<script lang="ts">
	// Track structures
	const personalityTrack = {
		title: 'Personality Development Track',
		desc: 'Designed to elevate students\' interpersonal qualities, leadership traits, and professional presentation capabilities to align with modern corporate and research expectations.',
		modules: [
			{ name: 'Grooming & Presentation', focus: 'Professional attire, body language, and corporate etiquette.' },
			{ name: 'Effective Communication', focus: 'Public speaking, active listening, and business writing.' },
			{ name: 'Teamwork & Collaboration', focus: 'Group dynamics, conflict resolution, and peer leadership.' },
			{ name: 'GD & Interview Readiness', focus: 'Mock group discussions, resume crafting, and technical interview simulations.' }
		]
	};

	const skillTrack = {
		title: 'Skill Building Track',
		desc: 'Hands-on technical advancement pathways structured around high-demand industry skills, software systems, and engineering frameworks.',
		modules: [
			{ name: 'Artificial Intelligence', focus: 'Fundamentals of machine learning, neural networks, and prompt engineering.' },
			{ name: 'Data Analytics & Stats', focus: 'Data visualization, SQL, python analytics, and business intelligence.' },
			{ name: 'Cyber Security Operations', focus: 'Network security, secure coding, threat intelligence, and ethical hacking.' },
			{ name: 'Project Management & Digital', focus: 'Agile methodologies, digital marketing campaigns, and systems administration.' }
		]
	};
</script>

<section id="tracks" class="py-16 sm:py-20 bg-white border-t border-slate-200">
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		
		<!-- Section Header -->
		<div class="pb-8 mb-12">
			<span class="text-[10px] font-bold tracking-[0.2em] text-[#881B1B] uppercase font-sans">
				Curriculum
			</span>
			<h2 class="text-2xl sm:text-3xl font-bold font-serif text-primary-dark tracking-tight mt-2">
				Student Advancement Tracks
			</h2>
			<p class="text-sm text-slate-500 font-sans max-w-2xl mt-3 leading-relaxed">
				ISPARC facilitates two core learning pipelines tailored to cultivate well-rounded professionals. These paths combine technical mastery with essential leadership and behavioral grooming.
			</p>
		</div>

		<!-- Tracks Grid (2 Columns) -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
			
			<!-- Track 1: Personality Development -->
			<div class="bg-[#F8F8F8] border border-slate-200 p-6 sm:p-8 rounded-sm">
				<div class="space-y-5">
					<div class="inline-flex items-center px-3 py-1 bg-[#881B1B] text-[10px] font-bold text-white uppercase tracking-wider rounded-sm">
						Track A
					</div>
					<h3 class="text-xl font-bold font-serif text-primary-dark tracking-tight">
						{personalityTrack.title}
					</h3>
					<p class="text-sm text-slate-600 leading-relaxed font-sans">
						{personalityTrack.desc}
					</p>
					
					<!-- Syllabus List -->
					<div class="space-y-3 border-t border-slate-200 pt-5">
						<h4 class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Syllabus Highlights</h4>
						<ul class="space-y-3">
							{#each personalityTrack.modules as mod}
								<li class="flex items-start gap-3">
									<span class="w-2 h-2 bg-[#C89B3C] mt-1.5 shrink-0 rounded-full"></span>
									<div>
										<span class="text-sm font-semibold text-slate-800 block">{mod.name}</span>
										<span class="text-xs text-slate-500 leading-relaxed block mt-0.5">{mod.focus}</span>
									</div>
								</li>
							{/each}
						</ul>
					</div>
				</div>
			</div>

			<!-- Track 2: Skill Building -->
			<div class="bg-[#F8F8F8] border border-slate-200 p-6 sm:p-8 rounded-sm">
				<div class="space-y-5">
					<div class="inline-flex items-center px-3 py-1 bg-[#881B1B] text-[10px] font-bold text-white uppercase tracking-wider rounded-sm">
						Track B
					</div>
					<h3 class="text-xl font-bold font-serif text-primary-dark tracking-tight">
						{skillTrack.title}
					</h3>
					<p class="text-sm text-slate-600 leading-relaxed font-sans">
						{skillTrack.desc}
					</p>
					
					<!-- Syllabus List -->
					<div class="space-y-3 border-t border-slate-200 pt-5">
						<h4 class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Syllabus Highlights</h4>
						<ul class="space-y-3">
							{#each skillTrack.modules as mod}
								<li class="flex items-start gap-3">
									<span class="w-2 h-2 bg-[#C89B3C] mt-1.5 shrink-0 rounded-full"></span>
									<div>
										<span class="text-sm font-semibold text-slate-800 block">{mod.name}</span>
										<span class="text-xs text-slate-500 leading-relaxed block mt-0.5">{mod.focus}</span>
									</div>
								</li>
							{/each}
						</ul>
					</div>
				</div>
			</div>

		</div>

	</div>
</section>
