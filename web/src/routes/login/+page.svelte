<script lang="ts">
  import HeaderSection from './HeaderSection.svelte';
  import LoginSection from './LoginSection.svelte';
  import FooterSection from './FooterSection.svelte';
</script>

<svelte:head>
  <title>Student Login | iSPARC IIPS DAVV</title>
  <meta name="description" content="Sign in to your iSPARC account - Institute of Professional Studies (IIPS), Devi Ahilya Vishwavidyalaya, Indore." />
</svelte:head>

<div class="min-h-screen bg-bg-main flex flex-col items-center w-full">
  <HeaderSection />
  <main class="flex-grow w-full flex justify-center items-start">
    <LoginSection />
  </main>
  <FooterSection />
</div>
