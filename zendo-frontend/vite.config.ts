import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig({
	plugins: [
		tailwindcss(), 
		sveltekit(),
		SvelteKitPWA({
			strategies: 'generateSW',
			registerType: 'autoUpdate',
			workbox: {
				globPatterns: ['client/**/*.{js,css,ico,png,svg,webp,webmanifest}'],
				cleanupOutdatedCaches: true,
				sourcemap: true,
				// Exclude problematic files from precaching
				globIgnores: [
					'**/node_modules/**/*',
					'**/sw.js',
					'**/workbox-*.js',
					'**/sw.js.map',
					'**/workbox-*.js.map'
				]
			},
			devOptions: {
				enabled: true,
				type: 'module'
			},
			manifest: {
				name: 'Zendo - Weekly Todo',
				short_name: 'Zendo',
				description: 'A beautiful weekly todo app for mindful productivity',
				theme_color: '#1f2937',
				background_color: '#1f2937',
				display: 'standalone',
				orientation: 'portrait',
				scope: '/',
				start_url: '/',
				icons: [
					{
						src: 'icon-48x48.png',
						sizes: '48x48',
						type: 'image/png'
					},
					{
						src: 'icon-72x72.png',
						sizes: '72x72',
						type: 'image/png'
					},
					{
						src: 'icon-96x96.png',
						sizes: '96x96',
						type: 'image/png'
					},
					{
						src: 'icon-128x128.png',
						sizes: '128x128',
						type: 'image/png'
					},
					{
						src: 'icon-144x144.png',
						sizes: '144x144',
						type: 'image/png'
					},
					{
						src: 'icon-152x152.png',
						sizes: '152x152',
						type: 'image/png'
					},
					{
						src: 'icon-192x192.png',
						sizes: '192x192',
						type: 'image/png'
					},
					{
						src: 'icon-256x256.png',
						sizes: '256x256',
						type: 'image/png'
					},
					{
						src: 'icon-384x384.png',
						sizes: '384x384',
						type: 'image/png'
					},
					{
						src: 'icon-512x512.png',
						sizes: '512x512',
						type: 'image/png'
					}
				]
			}
		})
	],
	define: {
		'process.env.NODE_ENV': process.env.NODE_ENV === 'production' 
			? '"production"'
			: '"development"'
	}
});
