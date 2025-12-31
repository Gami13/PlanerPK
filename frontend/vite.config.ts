import stylex from "@stylexjs/unplugin";
import devtools from "solid-devtools/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
export default defineConfig({
	plugins: [
		// StyleX must be before framework plugins for proper Fast Refresh
		stylex.vite({
			useCSSLayers: true,
			dev: process.env.NODE_ENV === "development",
			runtimeInjection: false,
		}),
		devtools(),
		solidPlugin(),
	],
	server: {
		port: 3000,
	},
	build: {
		outDir: "../static/",
		target: "esnext",
	},
});
