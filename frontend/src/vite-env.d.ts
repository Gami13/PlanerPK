/// <reference types="vite/client" />

// StyleX virtual CSS module for dev HMR
declare module "/virtual:stylex.css";

declare global {
	interface Window {
		__RUNTIME_CONFIG__: {
			API_BASE_URL: string;
		};
	}
}

export {};
