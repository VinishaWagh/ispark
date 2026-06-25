<script lang="ts">
  const navItems = [
    {
      label: "Home",
      href: "/",
    },
  ];
</script>

<header class="w-full bg-white border-b border-border-base sticky top-0 z-50 shadow-xs">
  <div class="max-w-6xl mx-auto flex items-center justify-between px-6 py-4">
    <!-- Brand Logo -->
    <a href="/" class="flex items-center gap-2 group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inst-navy focus-visible:ring-offset-2 rounded-lg transition-all duration-200">
      <div class="flex flex-col justify-center">
        <div class="flex items-baseline text-2xl font-bold tracking-tight text-slate-900 leading-none">
          <span>i</span><span class="text-acad-red">SPARC</span>
        </div>
      </div>
    </a>
    
    <!-- Navigation -->
    <nav aria-label="Primary" class="flex items-center gap-7">
      {#each navItems as item}
        <a
          href={item.href}
          aria-current={item.label === "Home" ? "page" : undefined}
          class="inline-flex items-center gap-1.5 focus:outline-none focus-visible:ring-2 focus-visible:ring-inst-navy focus-visible:ring-offset-2 rounded-md px-2 py-1 text-slate-650 hover:text-inst-navy transition-all duration-200"
        >
          <!-- Home Icon (inline SVG) -->
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4" aria-hidden="true">
            <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
            <polyline points="9 22 9 12 15 12 15 22"/>
          </svg>
          <span class="text-[13px] font-semibold uppercase tracking-wider">
            {item.label}
          </span>
        </a>
      {/each}
    </nav>
  </div>
</header>
