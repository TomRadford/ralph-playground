<script lang="ts">
	import { browser } from "$app/environment";
	import { page } from "$app/state";
	import favicon from "$lib/assets/favicon.svg";
	import { configureApiClient } from "$lib/configure-api-client";
	import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";

	let { children } = $props();

	configureApiClient();

	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser
			}
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<QueryClientProvider client={queryClient}>
	<div class="app-shell">
		<header class="topbar">
			<div>
				<p class="eyebrow">Ralph Playground</p>
				<h1>Office Foosball</h1>
			</div>
			<nav aria-label="Primary">
				<a
					href="/"
					class="nav-link"
					class:active={page.url.pathname === "/"}>Overview</a
				>
				<a
					href="/players"
					class="nav-link"
					class:active={page.url.pathname.startsWith("/players")}
					>Players</a
				>
				<a
					href="/matches"
					class="nav-link"
					class:active={page.url.pathname.startsWith("/matches")}
					>Matches</a
				>
				<a
					href="/leaderboard"
					class="nav-link"
					class:active={page.url.pathname.startsWith("/leaderboard")}
					>Leaderboard</a
				>
			</nav>
		</header>
		<main class="page">
			{@render children()}
		</main>
	</div>
</QueryClientProvider>

<style>
	:global(body) {
		margin: 0;
		background: linear-gradient(180deg, #eef6ff 0%, #f7f9fc 45%, #ffffff 100%);
		color: #1f2a37;
		font-family:
			Inter,
			system-ui,
			-apple-system,
			BlinkMacSystemFont,
			"Segoe UI",
			sans-serif;
	}

	.app-shell {
		margin: 0 auto;
		max-width: 70rem;
		padding: 1rem 1rem 1.5rem;
	}

	.topbar {
		padding: 1rem 1.1rem;
		border-radius: 16px;
		background: #0b2447;
		color: #f8fbff;
		box-shadow: 0 8px 30px rgba(11, 36, 71, 0.22);
		display: grid;
		gap: 0.9rem;
	}

	.eyebrow {
		margin: 0;
		font-size: 0.75rem;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		opacity: 0.75;
	}

	h1 {
		margin: 0.35rem 0 0;
		font-size: clamp(1.45rem, 2.2vw, 2rem);
	}

	nav {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.nav-link {
		color: #c6ddff;
		text-decoration: none;
		font-weight: 600;
		font-size: 0.95rem;
		padding: 0.45rem 0.7rem;
		border-radius: 9px;
		border: 1px solid transparent;
	}

	.nav-link:hover {
		background: rgba(255, 255, 255, 0.12);
		border-color: rgba(255, 255, 255, 0.16);
	}

	.nav-link.active {
		background: #f8fbff;
		color: #0b2447;
	}

	.page {
		margin-top: 1rem;
	}
</style>
