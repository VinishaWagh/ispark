<script lang="ts">
	import { API_BASE_URL } from '$lib/config';

	// Types
	interface CategoryProgress {
		name: string;
		credits: number;
		percentage: number;
	}

	interface SemesterRecord {
		name: string;
		credits: number;
		activitiesCount: number;
		grade: string;
		gradeColorClass: string;
	}

	interface GradeScale {
		name: string;
		range: string;
		isActive: boolean;
	}

	interface Insight {
		text: string;
		bgClass: string;
		textClass: string;
		borderClass: string;
	}

	let token = localStorage.getItem('access_token') || '';
	let totalCredits = $state(0);
	let categoryProgresses = $state<CategoryProgress[]>([]);
	let semesterRecords = $state<SemesterRecord[]>([]);
	let loading = $state(true);

	async function loadProgressData() {
		try {
			const res = await fetch(`${API_BASE_URL}/api/student/marksheet`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (res.ok) {
				const data = await res.json();
				totalCredits = data.total_credits;

				categoryProgresses = (data.credit_categories || []).map((cat: any) => {
					let req = 40; // baseline
					if (cat.category.toLowerCase().includes('technical')) req = 60;
					else if (cat.category.toLowerCase().includes('social')) req = 30;

					return {
						name: cat.category,
						credits: cat.credits,
						percentage: Math.min(Math.round((cat.credits / req) * 100), 100)
					};
				});

				semesterRecords = (data.semester_summary || []).map((sem: any) => {
					let grade = 'Grade D';
					let color = 'text-slate-500';
					if (sem.credits >= 30) {
						grade = 'Grade O';
						color = 'text-emerald-600';
					} else if (sem.credits >= 20) {
						grade = 'Grade A';
						color = 'text-[#881B1B]';
					} else if (sem.credits >= 15) {
						grade = 'Grade B';
						color = 'text-blue-600';
					} else if (sem.credits >= 10) {
						grade = 'Grade C';
						color = 'text-amber-500';
					}

					return {
						name: sem.semester,
						credits: sem.credits,
						activitiesCount: sem.activities,
						grade: grade,
						gradeColorClass: color
					};
				});
			}
		} catch (err) {
			console.error('Error fetching progress data:', err);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadProgressData();
	});

	// Derived statistics
	let percentComplete = $derived(Math.min(Math.round((totalCredits / 200) * 100), 100));
	let remainingCredits = $derived(Math.max(200 - totalCredits, 0));
	let currentGrade = $derived.by(() => {
		if (totalCredits >= 140) return 'O';
		if (totalCredits >= 100) return 'A';
		if (totalCredits >= 70) return 'B';
		if (totalCredits >= 40) return 'C';
		return 'D';
	});
	let nextGradeCredits = $derived.by(() => {
		if (totalCredits >= 140) return 0;
		if (totalCredits >= 100) return 140 - totalCredits;
		if (totalCredits >= 70) return 100 - totalCredits;
		if (totalCredits >= 40) return 70 - totalCredits;
		return 40 - totalCredits;
	});
	let nextGradeName = $derived.by(() => {
		if (totalCredits >= 140) return 'Max';
		if (totalCredits >= 100) return 'O';
		if (totalCredits >= 70) return 'A';
		if (totalCredits >= 40) return 'B';
		return 'C';
	});

	// Grade Scales active states computed dynamically from current totalCredits
	let gradeScales = $derived<GradeScale[]>([
		{ name: 'Grade O', range: '140+', isActive: currentGrade === 'O' },
		{ name: 'Grade A', range: '100-139', isActive: currentGrade === 'A' },
		{ name: 'Grade B', range: '70-99', isActive: currentGrade === 'B' },
		{ name: 'Grade C', range: '40-69', isActive: currentGrade === 'C' },
		{ name: 'Grade D', range: 'Below 40', isActive: currentGrade === 'D' }
	]);

	// Insights calculated dynamically
	let insights = $derived.by<Insight[]>(() => {
		const list: Insight[] = [];
		// Find highest category
		if (categoryProgresses.length > 0) {
			const highest = [...categoryProgresses].sort((a, b) => b.credits - a.credits)[0];
			list.push({
				text: `Your strongest contribution area is ${highest.name}.`,
				bgClass: 'bg-emerald-50/70',
				textClass: 'text-emerald-800',
				borderClass: 'border-emerald-150'
			});
		}

		if (totalCredits < 140) {
			list.push({
				text: `You need ${140 - totalCredits} more credits to reach Grade O.`,
				bgClass: 'bg-rose-50/70',
				textClass: 'text-rose-800',
				borderClass: 'border-rose-150'
			});
		} else {
			list.push({
				text: 'Excellent! You have unlocked Grade O, the highest grade level.',
				bgClass: 'bg-amber-50/70',
				textClass: 'text-amber-800',
				borderClass: 'border-amber-150'
			});
		}

		list.push({
			text: 'A minimum of 40 credits is required to clear the extracurricular criteria.',
			bgClass: 'bg-blue-50/70',
			textClass: 'text-blue-800',
			borderClass: 'border-blue-150'
		});

		return list;
	});
</script>

<div class="space-y-6">
	{#if loading}
		<div class="bg-white border border-slate-200 rounded-xl p-12 text-center shadow-xs flex items-center justify-center">
			<svg class="animate-spin h-6 w-6 text-[#881B1B]" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
			<span class="ml-3 text-xs font-bold text-slate-500 uppercase tracking-widest animate-pulse">Loading Progress Data...</span>
		</div>
	{:else}
		<!-- ==================== 1. OVERVIEW STAT BOX ==================== -->
		<section class="bg-white border border-slate-200 p-6 rounded-xl shadow-xs space-y-6">
			<div class="flex flex-col md:flex-row md:items-start justify-between gap-6">
				<div>
					<span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider"
						>Academic Year 2025-26</span
					>
					<div class="flex items-baseline gap-2 mt-1">
						<span class="text-5xl font-extrabold font-serif text-[#0B1535] leading-none">{totalCredits}</span>
						<span class="text-xs font-bold text-slate-405 uppercase tracking-widest"
							>Credits Earned</span
						>
					</div>
				</div>

				<!-- Grid of Sub-stats -->
				<div class="grid grid-cols-3 gap-6 md:gap-12 text-slate-800 shrink-0">
					<div class="flex flex-col">
						<span class="text-[9px] font-bold text-slate-405 uppercase tracking-wider">Target</span>
						<span class="text-lg font-bold font-serif text-[#0B1535] mt-1">200</span>
						<span class="text-[8px] font-bold text-slate-405 uppercase tracking-wide"
							>Credits Required</span
						>
					</div>
					<div class="flex flex-col">
						<span class="text-[9px] font-bold text-slate-405 uppercase tracking-wider"
							>Current Grade</span
						>
						<span class="text-lg font-bold font-serif text-[#881B1B] mt-1">{currentGrade}</span>
						<span class="text-[8px] font-bold text-slate-405 uppercase tracking-wide">&nbsp;</span>
					</div>
					<div class="flex flex-col">
						<span class="text-[9px] font-bold text-slate-405 uppercase tracking-wider"
							>{totalCredits >= 140 ? 'Status' : `To Grade ${nextGradeName}`}</span
						>
						<span class="text-lg font-bold font-serif text-[#0B1535] mt-1">{totalCredits >= 140 ? 'Maxed' : nextGradeCredits}</span>
						<span class="text-[8px] font-bold text-slate-405 uppercase tracking-wide"
							>{totalCredits >= 140 ? 'Out of Goals' : 'Credits Needed'}</span
						>
					</div>
				</div>
			</div>

			<!-- Progress bar block -->
			<div class="space-y-2">
				<div class="flex justify-between text-xs font-bold text-slate-700">
					<span>{totalCredits} / 200 Credits Completed</span>
					<span>{percentComplete}%</span>
				</div>

				<div class="h-3 w-full bg-slate-100 rounded-full overflow-hidden relative">
					<div class="h-full bg-[#881B1B] rounded-full" style="width: {percentComplete}%"></div>
				</div>

				<!-- Labels scale -->
				<div class="flex justify-between text-[10px] font-bold text-slate-405 font-sans px-1">
					<span>0</span>
					<span>50</span>
					<span>100</span>
					<span>150</span>
					<span>200</span>
				</div>
			</div>
		</section>

		<!-- ==================== 2. DOUBLE COLUMN CONTENT ==================== -->
		<section class="grid grid-cols-1 lg:grid-cols-12 gap-6">
			<!-- LEFT COLUMN -->
			<div class="lg:col-span-8 space-y-6">
				<!-- Credit Distribution Card -->
				<div class="bg-white border border-slate-200 p-6 rounded-xl shadow-xs">
					<h2 class="text-base font-bold font-serif text-[#0B1535] pb-4 border-b border-slate-100">
						Credit Distribution by Activity Category
					</h2>

					<div class="space-y-4.5 mt-5">
						{#if categoryProgresses.length > 0}
							{#each categoryProgresses as cat}
								<div class="space-y-1.5">
									<div class="flex justify-between text-xs font-bold text-slate-800">
										<span class="text-[11px] font-bold text-slate-655">{cat.name}</span>
										<span>{cat.credits} credits</span>
									</div>
									<div class="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
										<div
											class="h-full bg-[#0B1535] rounded-full"
											style="width: {cat.percentage}%"
										></div>
									</div>
								</div>
							{/each}
						{:else}
							<p class="text-xs text-slate-450 font-semibold text-center py-6">No credit category stats available. Upload certificates to earn credits.</p>
						{/if}
					</div>
				</div>

				<!-- Semester Progress Card -->
				<div class="bg-white border border-slate-200 p-5 rounded-xl shadow-xs">
					<h2 class="text-base font-bold font-serif text-[#0B1535] pb-4 border-b border-slate-100">
						Semester Progress
					</h2>

					<div class="overflow-x-auto mt-4">
						<table class="w-full text-left border-collapse text-xs">
							<thead>
								<tr
									class="text-[10px] font-bold text-[#6B7280] uppercase tracking-widest border-b border-slate-100 bg-slate-50/50"
								>
									<th class="py-3 px-4">Semester</th>
									<th class="py-3 px-4">Credits Earned</th>
									<th class="py-3 px-4">Activities</th>
									<th class="py-3 px-4 text-right">Grade Contribution</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-100">
								{#each semesterRecords as sem}
									<tr class="hover:bg-slate-50/50 transition-colors">
										<td class="py-3.5 px-4 font-bold text-[#0B1535]">{sem.name}</td>
										<td class="py-3.5 px-4 text-slate-700 font-semibold">{sem.credits} credits</td>
										<td class="py-3.5 px-4 text-slate-500">{sem.activitiesCount} Activities</td>
										<td class="py-3.5 px-4 text-right font-bold {sem.gradeColorClass}">{sem.grade}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			</div>

			<!-- RIGHT COLUMN -->
			<div class="lg:col-span-4 space-y-6">
				<!-- Current Grade Standing Card -->
				<div class="bg-white border border-slate-200 p-6 rounded-xl shadow-xs">
					<h2 class="text-base font-bold font-serif text-[#0B1535] pb-4 border-b border-slate-100">
						Current Grade Standing
					</h2>

					<!-- Active Grade Banner Block -->
					<div
						class="mt-4 p-4.5 bg-[#881B1B]/5 border border-[#881B1B]/15 rounded-xl flex items-center gap-4"
					>
						<div
							class="w-14 h-14 rounded-lg bg-[#881B1B] text-white flex items-center justify-center font-bold text-2xl font-serif shrink-0 border border-[#881B1B]/20 shadow-xs"
						>
							{currentGrade}
						</div>
						<div>
							<span class="text-[10px] font-bold text-[#881B1B] uppercase tracking-wider block"
								>Active standing</span
							>
							<span class="text-xs font-bold text-[#0B1535] leading-tight block mt-0.5"
								>Level Achieved</span
							>
							<span class="text-[10px] text-slate-405 font-medium block mt-1 leading-normal"
								>Verified extracurricular credentials evaluated.</span
							>
						</div>
					</div>

					<!-- Grade Scale List -->
					<div class="space-y-2 mt-5">
						{#each gradeScales as scale}
							<div
								class="flex items-center justify-between p-3 rounded-lg border text-xs transition duration-200 {scale.isActive
									? 'bg-[#881B1B]/5 border-[#881B1B] text-[#881B1B] font-bold ring-1 ring-[#881B1B]/10'
									: 'bg-white border-slate-150 hover:border-slate-300 text-slate-700'}"
							>
								<span>{scale.name}</span>
								<span>{scale.range} credits</span>
							</div>
						{/each}
					</div>
				</div>

				<!-- Mentorship Guidance Insights -->
				<div class="bg-white border border-slate-200 p-6 rounded-xl shadow-xs space-y-4">
					<h2 class="text-base font-bold font-serif text-[#0B1535] pb-3 border-b border-slate-100">
						Guidance & Insights
					</h2>

					<div class="space-y-3.5">
						{#each insights as insight}
							<div
								class="p-4 border rounded-xl text-xs font-medium leading-relaxed font-sans {insight.bgClass} {insight.borderClass} {insight.textClass}"
							>
								{insight.text}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</section>
	{/if}
</div>
