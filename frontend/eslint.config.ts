import js from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import solidPlugin from "eslint-plugin-solid";
import { defineConfig } from "eslint/config";
import globals from "globals";
import tseslint from "typescript-eslint";

export default defineConfig(
	js.configs.recommended,
	...tseslint.configs.recommended,
	{
		files: ["**/*.{ts,tsx,js,jsx}"],
		languageOptions: {
			globals: globals.browser,
			parserOptions: {
				ecmaFeatures: {
					jsx: true,
				},
			},
		},
	},
	{
		files: ["**/*.{ts,tsx}"],
		plugins: {
			// @ts-expect-error - eslint-plugin-solid doesn't have proper flat config types yet
			solid: solidPlugin,
		},
		rules: {
			...solidPlugin.configs.recommended.rules,
		},
	},
	eslintConfigPrettier
);
