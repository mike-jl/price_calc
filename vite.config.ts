import { defineConfig } from 'vite';

export default defineConfig({
    build: {
        rollupOptions: {
            input: './scripts/main.ts', //  Tell Vite what your entry point is
        },
        outDir: './assets/js', //  Where the output .js file goes
        emptyOutDir: false,    // Don't wipe your assets dir if you're mixing files
        manifest: true, //  Create a manifest.json file
    }
});

