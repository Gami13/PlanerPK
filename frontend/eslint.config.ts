import js from "@eslint/js";
import stylexPlugin from "@stylexjs/eslint-plugin";
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
			// @ts-expect-error - @stylexjs/eslint-plugin doesn't have proper flat config types yet
			"@stylexjs": stylexPlugin,
		},
		rules: {
			...solidPlugin.configs.recommended.rules,
			// Validates StyleX styles (syntax, properties, values)
			"@stylexjs/valid-styles": "error",
			// Sorts style keys alphabetically for consistency
			"@stylexjs/sort-keys": "warn",
			// Requires shorthand properties to be split into individual properties
			"@stylexjs/valid-shorthands": "warn",
			// Detects unused StyleX styles
			"@stylexjs/no-unused": "warn",
			// Enforces standard CSS properties and values
			"@stylexjs/no-nonstandard-styles": "error",
			// Requires theme/constant files to use .stylex extension
			"@stylexjs/enforce-extension": "error",
			// Disallows legacy contextual style patterns
			"@stylexjs/no-legacy-contextual-styles": "error",
			// Warns about :has() selectors with limited browser support
			"@stylexjs/no-lookahead-selectors": "warn",
		},
	},
	eslintConfigPrettier
);
