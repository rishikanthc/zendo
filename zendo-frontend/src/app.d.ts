// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

// PWA Virtual Module Types
declare module 'virtual:pwa-info' {
	export interface PWAInfo {
		webManifest: {
			linkTag: string;
		};
	}
	export const pwaInfo: PWAInfo | undefined;
}

declare module 'virtual:pwa-register' {
	export interface RegisterSWOptions {
		immediate?: boolean;
		onRegistered?: (registration: ServiceWorkerRegistration | undefined) => void;
		onRegisterError?: (error: any) => void;
		onNeedRefresh?: () => void;
		onOfflineReady?: () => void;
		onUpdateFound?: () => void;
	}
	
	export function registerSW(options?: RegisterSWOptions): (reloadPage?: boolean) => Promise<void>;
}

declare module 'virtual:pwa-register/svelte' {
	import type { Writable } from 'svelte/store';
	
	export interface UseRegisterSWOptions {
		immediate?: boolean;
		onRegistered?: (registration: ServiceWorkerRegistration | undefined) => void;
		onRegisterError?: (error: any) => void;
		onNeedRefresh?: () => void;
		onOfflineReady?: () => void;
		onUpdateFound?: () => void;
	}
	
	export interface UseRegisterSWReturn {
		needRefresh: Writable<boolean>;
		offlineReady: Writable<boolean>;
		updateServiceWorker: (reloadPage?: boolean) => Promise<void>;
	}
	
	export function useRegisterSW(options?: UseRegisterSWOptions): UseRegisterSWReturn;
}

export {};
