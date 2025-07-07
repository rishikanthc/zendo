<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import Toaster from '$lib/components/ui/sonner/sonner.svelte';

	let { children } = $props();

	onMount(async () => {
		try {
			const { pwaInfo } = await import('virtual:pwa-info');
			if (pwaInfo) {
				const { registerSW } = await import('virtual:pwa-register');
				registerSW({
					immediate: true,
					onRegistered(r: any) {
						console.log(`SW Registered: ${r}`);
					},
					onRegisterError(error: any) {
						console.log('SW registration error', error);
					}
				});
			}
		} catch (error) {
			console.log('PWA not available:', error);
		}
	});
</script>

<svelte:head>
	<meta name="theme-color" content="#1f2937" />
	<meta name="apple-mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-status-bar-style" content="default" />
	<meta name="apple-mobile-web-app-title" content="Zendo" />
	<meta name="mobile-web-app-capable" content="yes" />
	<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
</svelte:head>

{@render children()}

{#await import('$lib/ReloadPrompt.svelte') then { default: ReloadPrompt}}
	<ReloadPrompt />
{/await}

<Toaster />
