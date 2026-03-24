<script lang="ts">
	import { browser } from "$app/environment";
	import favicon from "$lib/assets/favicon.svg";
	import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
	import { configureApiClient } from "$lib/api/configure-client";

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
	{@render children()}
</QueryClientProvider>
