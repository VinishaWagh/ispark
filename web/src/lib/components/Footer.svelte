<script lang="ts">
	const year = new Date().getFullYear();
	const currentYear = year < 2026 ? 2026 : year;
</script>

<footer class="py-10 bg-[#F8F8F8] border-t border-slate-200">
	<div class="max-w-7xl mx-auto px-6 lg:px-8">
		<div class="flex flex-col md:flex-row items-center justify-between gap-4">
			<div class="flex items-center gap-2">
				<div class="flex items-baseline text-lg font-bold tracking-tight">
					<span class="text-slate-900">i</span><span class="text-[#881B1B]">SPARC</span>
				</div>
				<span class="text-xs text-slate-400">|</span>
				<span class="text-xs text-slate-500">IIPS DAVV CELL</span>
			</div>
			<p class="text-xs text-slate-500 font-sans">
				© {currentYear} IIPS DAVV. All rights reserved.
			</p>
		</div>
	</div>
</footer>
