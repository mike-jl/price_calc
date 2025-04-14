import js from '@eslint/js';
import * as tseslint from 'typescript-eslint';
import globals from 'globals';
import vitestGlobals from "eslint-config-vitest-globals/flat"

/** @type {import("eslint").Linter.FlatConfig[]} */
export default [
    js.configs.recommended,
    vitestGlobals(),

    {
        files: ['scripts/**/*.ts'],
        languageOptions: {
            parser: tseslint.parser,
            parserOptions: {
                project: './tsconfig.json',
                tsconfigRootDir: import.meta.dirname,
            },
            globals: {
                ...globals.browser,
                ...globals.es2021,
            },
        },
        plugins: {
            '@typescript-eslint': tseslint.plugin,
        },
        rules: {
            semi: ['error', 'always'],
            quotes: ['error', 'single'],
            "no-unused-vars": "off",
            "@typescript-eslint/no-unused-vars": ["error"],
            '@typescript-eslint/consistent-indexed-object-style': ['error', 'record'],
            '@typescript-eslint/consistent-type-definitions': ['error', 'interface'],
            '@typescript-eslint/no-explicit-any': 'warn',
        },
    },
];

