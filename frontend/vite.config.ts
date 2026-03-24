import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			"/health": "http://localhost:8080",
			"/players": "http://localhost:8080",
			"/matches": "http://localhost:8080",
			"/leaderboard": "http://localhost:8080"
		}
	}
});
