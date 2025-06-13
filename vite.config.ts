import tailwindcss from "@tailwindcss/vite";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { SvelteKitPWA } from "@vite-pwa/sveltekit";

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			manifest: {
				name: "Weeko",
				short_name: "Weeko",
				description: "Self-hosted weekly planner",
				theme_color: "#FDF6E3",
				background_color: "#FDF6E3",
				display: "standalone",
				start_url: "/",
				icons: [
					{
						src: "/icon-48x48.png",
						sizes: "48x48",
						type: "image/png",
					},
					{
						src: "/icon-72x72.png",
						sizes: "72x72",
						type: "image/png",
					},
					{
						src: "/icon-96x96.png",
						sizes: "96x96",
						type: "image/png",
					},
					{
						src: "/icon-128x128.png",
						sizes: "128x128",
						type: "image/png",
					},
					{
						src: "/icon-144x144.png",
						sizes: "144x144",
						type: "image/png",
					},
					{
						src: "/icon-152x152.png",
						sizes: "152x152",
						type: "image/png",
					},
					{
						src: "/icon-192x192.png",
						sizes: "192x192",
						type: "image/png",
					},
					{
						src: "/icon-256x256.png",
						sizes: "256x256",
						type: "image/png",
					},
					{
						src: "/icon-384x384.png",
						sizes: "384x384",
						type: "image/png",
					},
					{
						src: "/icon-512x512.png",
						sizes: "512x512",
						type: "image/png",
					},
				],
			},
			registerType: "autoUpdate",
		}),
	],
});
