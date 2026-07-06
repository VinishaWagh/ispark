<script lang="ts">
	import { fade } from 'svelte/transition';

	import { API_BASE_URL } from '$lib/config';

	let token = localStorage.getItem('access_token') || '';
	let studentInfo = $state({
		name: '',
		rollNo: '',
		enrollmentNo: '',
		semester: '',
		course: '',
		department: '',
		batch: '',
		institute: ''
	});
	let creditCategories = $state<any[]>([]);
	let totalActivities = $state(0);
	let totalCredits = $state(0);
	let totalContribution = $state('0%');
	let semesterSummary = $state<any[]>([]);
	let loading = $state(true);

	async function loadMarksheet() {
		try {
			const res = await fetch(`${API_BASE_URL}/api/student/marksheet`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (res.ok) {
				const data = await res.json();
				studentInfo = {
					name: data.student_info.name,
					rollNo: data.student_info.roll_no,
					enrollmentNo: data.student_info.enrollment_no,
					semester: data.student_info.semester,
					course: data.student_info.course,
					department: data.student_info.department,
					batch: data.student_info.batch,
					institute: data.student_info.institute
				};
				creditCategories = data.credit_categories || [];
				totalActivities = data.total_activities;
				totalCredits = data.total_credits;
				totalContribution = '100%';
				semesterSummary = data.semester_summary || [];
			}
		} catch (err) {
			console.error('Error fetching marksheet:', err);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadMarksheet();
	});

	// Toast State
	let toastMessage = $state('');
	let toastType = $state<'success' | 'info' | 'error'>('info');
	let showToast = $state(false);

	function triggerToast(message: string, type: 'success' | 'info' | 'error' = 'info') {
		toastMessage = message;
		toastType = type;
		showToast = true;
		setTimeout(() => {
			showToast = false;
		}, 3000);
	}

	// Action triggers
	function handleDownload() {
		triggerToast('Generating official PDF transcript...', 'info');
		setTimeout(() => {
			window.print();
		}, 1000);
	}

	function handlePrint() {
		window.print();
	}

	function handleShare() {
		const shareUrl = `${window.location.origin}${window.location.pathname}#verify/MARK-2026-001245`;
		navigator.clipboard.writeText(shareUrl).then(
			() => {
				triggerToast('Verification link copied to clipboard!', 'success');
			},
			() => {
				triggerToast('Failed to copy link. Please copy manually.', 'error');
			}
		);
	}
</script>

<!-- Toast Notification -->
{#if showToast}
	<div
		transition:fade={{ duration: 150 }}
		class="fixed bottom-5 right-5 z-[100] flex items-center gap-2.5 px-4 py-3 rounded-lg shadow-lg border text-xs font-semibold no-print transition-all duration-200 {toastType ===
		'success'
			? 'bg-emerald-50 text-emerald-800 border-emerald-200'
			: toastType === 'error'
				? 'bg-rose-50 text-rose-800 border-rose-200'
				: 'bg-[#0B1535] text-white border-slate-700'}"
	>
		{#if toastType === 'success'}
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 20 20"
				fill="currentColor"
				class="w-4 h-4 text-emerald-500"
			>
				<path
					fill-rule="evenodd"
					d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16zm3.857-9.809a.75.75 0 0 0-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 1 0-1.06 1.061l2.5 2.5a.75.75 0 0 0 1.137-.089l4-5.5z"
					clip-rule="evenodd"
				/>
			</svg>
		{:else if toastType === 'error'}
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 20 20"
				fill="currentColor"
				class="w-4 h-4 text-rose-500"
			>
				<path
					fill-rule="evenodd"
					d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM8.28 7.22a.75.75 0 0 0-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 1 0 1.06 1.06L10 11.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L11.06 10l1.72-1.72a.75.75 0 0 0-1.06-1.06L10 8.94 8.28 7.22z"
					clip-rule="evenodd"
				/>
			</svg>
		{:else}
			<svg class="animate-spin h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
				></circle>
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				></path>
			</svg>
		{/if}
		{toastMessage}
	</div>
{/if}

<div class="space-y-6 font-sans">
	<!-- Action buttons row -->
	<div class="flex flex-wrap items-center justify-end gap-3.5 no-print">
		<!-- Download PDF button -->
		<button
			onclick={handleDownload}
			class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#0B1535] hover:bg-[#1a2b5e] text-white font-extrabold text-[11px] tracking-wider uppercase rounded shadow-sm hover:shadow transition duration-200 cursor-pointer"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2.5"
				stroke="currentColor"
				class="w-4 h-4"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
				/>
			</svg>
			Download PDF
		</button>

		<!-- Print Marksheet button -->
		<button
			onclick={handlePrint}
			class="inline-flex items-center gap-2 px-5 py-2.5 bg-white hover:bg-slate-50 text-slate-800 font-extrabold text-[11px] tracking-wider uppercase rounded border border-slate-250 shadow-2xs hover:border-slate-350 transition duration-200 cursor-pointer"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2.2"
				stroke="currentColor"
				class="w-4 h-4"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M6.72 13.844 2.44 11c0-.12-.007-.24-.007-.36 0-3.233 3.007-5.855 6.716-5.855 3.513 0 6.446 2.372 6.702 5.402.012.152.018.307.018.462a4.42 4.42 0 0 1-.007.36l-4.28 2.844m-6.72 0a5.976 5.976 0 0 0 6.72 0m-6.72 0H3.36a1.68 1.68 0 0 0-1.68 1.68v2.52a1.68 1.68 0 0 0 1.68 1.68h11.28a1.68 1.68 0 0 0 1.68-1.68v-2.52a1.68 1.68 0 0 0-1.68-1.68h-2.24m-6.72 0v2.24m6.72-2.24v2.24M8 19.6h4"
				/>
			</svg>
			Print Marksheet
		</button>

		<!-- Share Record button -->
		<button
			onclick={handleShare}
			class="inline-flex items-center gap-2 px-5 py-2.5 bg-white hover:bg-slate-50 text-slate-800 font-extrabold text-[11px] tracking-wider uppercase rounded border border-slate-250 shadow-2xs hover:border-slate-350 transition duration-200 cursor-pointer"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2.2"
				stroke="currentColor"
				class="w-4 h-4"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186.003-.004c.18-.326.28-.7-.003-2.182m0 2.186 9.566 5.314m-9.566-5.314 9.563-5.314m0 0a2.25 2.25 0 1 0 3.935-2.186 2.25 2.25 0 0 0-3.935 2.186Zm0-2.186-.003.004-.007-.004Zm9.566 12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z"
				/>
			</svg>
			Share Record
		</button>
	</div>

	<!-- Marksheet Card Container -->
	<article
		class="bg-white border border-slate-200 rounded-xl p-8 sm:p-12 shadow-sm max-w-4xl mx-auto w-full marksheet-card relative"
	>
		<!-- Transcript Blue Header Bar -->
		<header
			class="bg-[#0B1535] text-white p-6 rounded-t flex flex-col md:flex-row justify-between items-center text-center md:text-left gap-4"
		>
			<div class="flex flex-col gap-1">
				<span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest leading-none"
					>OFFICIAL DOCUMENT</span
				>
				<span class="text-xs font-semibold text-slate-200 mt-1">Academic Year 2025–26</span>
			</div>

			<div class="flex flex-col items-center">
				<div class="flex flex-col items-center">
					<!-- iSPARC Logo Text styled exactly like sidebar -->
					<h2 class="text-2xl font-black tracking-widest text-white font-serif">
						i<span class="text-red-500">SPARC</span>
					</h2>
					<span
						class="text-[9px] font-extrabold tracking-wider text-slate-300 uppercase -mt-1 block"
						>Skill, Personality Advancement & Refinement Cell</span
					>
					<span
						class="text-[8px] font-bold tracking-normal text-slate-400 mt-0.5 text-center leading-normal"
					>
						International Institute of Professional Studies (IIPS)<br />
						Devi Ahilya Vishwavidyalaya, Indore
					</span>
				</div>
			</div>

			<div class="flex flex-col md:items-end gap-1">
				<span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest leading-none"
					>MARKSHEET TYPE</span
				>
				<span class="text-xs font-semibold text-slate-200 mt-1">Extracurricular Transcript</span>
			</div>
		</header>

		<!-- Main Marksheet Content -->
		<div class="py-6 space-y-8">
			<!-- Section Title: OFFICIAL EXTRACURRICULAR MARKSHEET -->
			<div class="text-center py-2.5 border-y-4 border-double border-slate-200">
				<h3 class="text-lg sm:text-xl font-extrabold text-[#0B1535] tracking-widest font-serif">
					OFFICIAL EXTRACURRICULAR MARKSHEET
				</h3>
			</div>

			<!-- STUDENT INFORMATION -->
			<section class="space-y-3" aria-labelledby="student-info-title">
				<h4
					id="student-info-title"
					class="text-[11px] font-bold text-slate-400 tracking-wider uppercase"
				>
					STUDENT INFORMATION
				</h4>
				<hr class="border-t border-slate-200" />

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-y-3 gap-x-12 text-xs pt-1.5">
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>STUDENT NAME</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.name}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>ENROLLMENT NUMBER</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.enrollmentNo}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>ROLL NUMBER</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.rollNo}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>COURSE</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.course}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>SEMESTER</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.semester}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>DEPARTMENT</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.department}</span>
					</div>
					<div class="flex justify-between border-b border-slate-50 pb-2 sm:border-b-0">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>BATCH</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.batch}</span>
					</div>
					<div class="flex justify-between pb-2 sm:border-b-0">
						<span class="text-slate-405 font-bold uppercase tracking-wider text-[10px] w-1/2"
							>INSTITUTE</span
						>
						<span class="font-bold text-[#0B1535] w-1/2">{studentInfo.institute}</span>
					</div>
				</div>
			</section>

			<!-- CREDIT DISTRIBUTION BY ACTIVITY CATEGORY -->
			<section class="space-y-4" aria-labelledby="credit-distribution-title">
				<div class="flex items-center justify-between">
					<h4
						id="credit-distribution-title"
						class="text-[11px] font-bold text-slate-400 tracking-wider uppercase"
					>
						CREDIT DISTRIBUTION BY ACTIVITY CATEGORY
					</h4>
				</div>
				<hr class="border-t border-slate-200" />

				<div class="overflow-x-auto">
					<table class="w-full text-left border-collapse text-xs">
						<thead>
							<tr
								class="bg-slate-50 text-[10px] font-bold text-slate-500 uppercase tracking-wider border-b border-slate-200"
							>
								<th class="py-3 px-4">Category</th>
								<th class="py-3 px-4 text-center">Activities</th>
								<th class="py-3 px-4 text-center">Credits Earned</th>
								<th class="py-3 px-4 text-right">Contribution</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 text-[#0B1535] font-bold">
							{#each creditCategories as item}
								<tr class="hover:bg-slate-50/50">
									<td class="py-3 px-4">{item.category}</td>
									<td class="py-3 px-4 text-center font-semibold text-slate-600"
										>{item.activities}</td
									>
									<td class="py-3 px-4 text-center font-serif text-slate-800 text-[13px]"
										>{item.credits}</td
									>
									<td class="py-3 px-4 text-right text-slate-500 font-semibold"
										>{item.contribution}</td
									>
								</tr>
							{/each}
							<!-- Totals Row -->
							<tr class="bg-slate-50 border-t-2 border-slate-300 text-sm">
								<td class="py-3.5 px-4 font-black uppercase tracking-wider">TOTAL</td>
								<td class="py-3.5 px-4 text-center font-black">{totalActivities}</td>
								<td class="py-3.5 px-4 text-center font-black font-serif text-base"
									>{totalCredits}</td
								>
								<td class="py-3.5 px-4 text-right font-black">{totalContribution}</td>
							</tr>
						</tbody>
					</table>
				</div>

				<!-- FINAL GRADE AWARDED BOX -->
				<div
					class="bg-rose-50/30 border border-rose-100 p-4 rounded flex items-center justify-between text-xs mt-4"
				>
					<span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest"
						>FINAL GRADE AWARDED</span
					>
					<span class="text-2xl font-serif font-black italic text-[#881B1B] pr-4">Grade A</span>
				</div>
			</section>

			<!-- SEMESTER CONTRIBUTION SUMMARY -->
			<section class="space-y-4" aria-labelledby="semester-summary-title">
				<h4
					id="semester-summary-title"
					class="text-[11px] font-bold text-slate-400 tracking-wider uppercase"
				>
					SEMESTER CONTRIBUTION SUMMARY
				</h4>
				<hr class="border-t border-slate-200" />

				<div class="overflow-x-auto">
					<table class="w-full text-left border-collapse text-xs">
						<thead>
							<tr
								class="bg-slate-50 text-[10px] font-bold text-slate-500 uppercase tracking-wider border-b border-slate-200"
							>
								<th class="py-3 px-4">Semester</th>
								<th class="py-3 px-4 text-center">Credits Earned</th>
								<th class="py-3 px-4 text-center">Activities Completed</th>
								<th class="py-3 px-4 text-right">Cumulative Credits</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 text-[#0B1535] font-bold">
							{#each semesterSummary as item}
								<tr class="hover:bg-slate-50/50">
									<td class="py-3 px-4">{item.semester}</td>
									<td class="py-3 px-4 text-center font-serif text-slate-800 text-[13px]"
										>{item.credits}</td
									>
									<td class="py-3 px-4 text-center font-semibold text-slate-600"
										>{item.activities}</td
									>
									<td class="py-3 px-4 text-right font-serif text-slate-800 text-[13px]"
										>{item.cumulative}</td
									>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>

			<!-- INSTITUTIONAL REMARKS -->
			<section class="space-y-3" aria-labelledby="remarks-title">
				<h4
					id="remarks-title"
					class="text-[11px] font-bold text-slate-400 tracking-wider uppercase"
				>
					INSTITUTIONAL REMARKS
				</h4>
				<hr class="border-t border-slate-200" />

				<div
					class="bg-slate-50/50 border border-slate-200/80 p-4 rounded text-xs text-slate-600 leading-relaxed font-medium"
				>
					The student has actively participated in extracurricular and co-curricular activities and
					has successfully fulfilled the requirements prescribed under the iSPARC framework. The
					records presented in this marksheet are based on verified certificates and approved
					activity logs maintained within the iSPARC Portal for Academic Year 2025-26.
				</div>
			</section>

			<!-- CERTIFICATION & APPROVAL -->
			<section class="space-y-6 pt-4" aria-labelledby="certification-title">
				<h4
					id="certification-title"
					class="text-[11px] font-bold text-slate-400 tracking-wider uppercase text-center md:text-left"
				>
					CERTIFICATION & APPROVAL
				</h4>
				<hr class="border-t border-slate-200" />

				<div class="grid grid-cols-1 md:grid-cols-3 items-end gap-8 pt-4">
					<!-- Faculty Mentor Sign -->
					<div class="text-center flex flex-col items-center">
						<span
							class="font-serif italic text-base text-[#0B1535] font-medium h-10 flex items-center justify-center"
							>Dr. A. Sharma</span
						>
						<hr class="border-t border-slate-350 w-44 my-2" />
						<span class="text-xs font-bold text-slate-800 block">Dr. A. Sharma</span>
						<span
							class="text-[10px] text-slate-400 font-semibold uppercase tracking-wider block mt-1"
							>Faculty Mentor, Dept. of Computer Applications</span
						>
					</div>

					<!-- QR verification block -->
					<div class="text-center flex flex-col items-center justify-center space-y-2">
						<div
							class="border border-slate-200 p-2 bg-white rounded shadow-3xs flex flex-col items-center gap-1"
						>
							<!-- Custom design high fidelity QR code SVG -->
							<svg class="w-16 h-16 text-[#0B1535]" viewBox="0 0 100 100" fill="currentColor">
								<!-- Finder Pattern top-left -->
								<rect x="0" y="0" width="28" height="28" />
								<rect x="4" y="4" width="20" height="20" fill="white" />
								<rect x="8" y="8" width="12" height="12" />

								<!-- Finder Pattern top-right -->
								<rect x="72" y="0" width="28" height="28" />
								<rect x="76" y="4" width="20" height="20" fill="white" />
								<rect x="80" y="8" width="12" height="12" />

								<!-- Finder Pattern bottom-left -->
								<rect x="0" y="72" width="28" height="28" />
								<rect x="4" y="76" width="20" height="20" fill="white" />
								<rect x="8" y="80" width="12" height="12" />

								<!-- QR random block patterns to simulate high resolution QR code -->
								<rect x="36" y="0" width="8" height="8" />
								<rect x="48" y="4" width="8" height="8" />
								<rect x="60" y="0" width="8" height="16" />

								<rect x="36" y="16" width="12" height="8" />
								<rect x="52" y="20" width="8" height="8" />
								<rect x="64" y="24" width="4" height="4" />

								<rect x="0" y="36" width="8" height="12" />
								<rect x="12" y="40" width="12" height="8" />
								<rect x="28" y="36" width="16" height="16" />
								<rect x="32" y="40" width="8" height="8" fill="white" />

								<rect x="48" y="36" width="8" height="8" />
								<rect x="60" y="40" width="16" height="8" />
								<rect x="84" y="36" width="16" height="12" />

								<rect x="0" y="52" width="12" height="8" />
								<rect x="16" y="56" width="8" height="8" />
								<rect x="28" y="56" width="12" height="12" />

								<rect x="48" y="52" width="16" height="16" />
								<rect x="52" y="56" width="8" height="8" fill="white" />
								<rect x="68" y="52" width="8" height="12" />
								<rect x="80" y="56" width="12" height="8" />

								<rect x="36" y="72" width="8" height="8" />
								<rect x="48" y="76" width="12" height="12" />
								<rect x="64" y="72" width="4" height="4" />

								<rect x="36" y="88" width="16" height="8" />
								<rect x="56" y="92" width="8" height="8" />

								<rect x="76" y="76" width="16" height="16" />
								<rect x="80" y="80" width="8" height="8" fill="white" />
							</svg>
							<span class="text-[8px] font-bold text-slate-400 tracking-wider uppercase mt-1"
								>Scan to verify</span
							>
						</div>
						<div class="text-[9px] text-slate-500 font-semibold tracking-wide">
							<p>Document ID: <strong class="text-slate-800">MARK-2026-001245</strong></p>
							<p class="mt-0.5">Generated: 24 Jun 2026</p>
						</div>
					</div>

					<!-- iSPARC Coordinator Sign -->
					<div class="text-center flex flex-col items-center">
						<span
							class="font-serif italic text-base text-[#0B1535] font-medium h-10 flex items-center justify-center font-serif"
							>Prof. R. Joshi</span
						>
						<hr class="border-t border-slate-350 w-44 my-2" />
						<span class="text-xs font-bold text-slate-800 block">Prof. R. Joshi</span>
						<span
							class="text-[10px] text-slate-400 font-semibold uppercase tracking-wider block mt-1"
							>iSPARC Coordinator, IIPS DAVV</span
						>
					</div>
				</div>

				<!-- Legal footer warning -->
				<div
					class="border border-slate-200 bg-slate-50/30 p-3 rounded text-[10px] text-slate-500 leading-relaxed font-semibold text-center mt-6"
				>
					This marksheet is generated based on verified extracurricular records maintained within
					the iSPARC Portal. Any discrepancy should be reported to the iSPARC Administration Office.
				</div>
			</section>
		</div>
	</article>
</div>

<style>
	/* Print specific media styling */
	@media print {
		/* Force margins and layout container rules */
		:global(aside),
		:global(header),
		.no-print {
			display: none !important;
		}

		:global(main) {
			padding: 0 !important;
			margin: 0 !important;
			background: white !important;
			width: 100% !important;
			max-width: 100% !important;
		}

		:global(body) {
			background-color: white !important;
			color: black !important;
		}

		.marksheet-card {
			border: none !important;
			box-shadow: none !important;
			padding: 0 !important;
			margin: 0 !important;
			width: 100% !important;
			max-width: 100% !important;
		}
	}
</style>
