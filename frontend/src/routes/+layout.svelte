<script lang="ts">
	import { browser } from "$app/environment";
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
	{@render children()}
</QueryClientProvider>
