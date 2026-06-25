<script lang="ts">
  const footerLines = [
    "Skill, Personality Advancement & Refinement Cell",
    "International Institute of Professional Studies (IIPS)",
    "Devi Ahilya Vishwavidyalaya (DAVV), Indore",
  ];
  
  const currentYear = new Date().getFullYear();
</script>

<footer class="w-full bg-inst-navy py-8 border-t border-slate-900 mt-auto">
  <div class="max-w-6xl mx-auto flex flex-col items-center gap-1.5 px-6 text-center">
    <h2 class="text-base font-bold text-white tracking-wider uppercase font-sans">
      iSPARC
    </h2>
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
