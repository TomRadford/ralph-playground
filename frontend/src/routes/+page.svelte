<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getLeaderboard, listMatches, listPlayers } from "$lib/api";

	const playersQuery = createQuery(() => ({
		queryKey: ["players"],
		queryFn: async () => {
			const response = await listPlayers();
			return response.data?.items ?? [];
		}
	}));

	const matchesQuery = createQuery(() => ({
		queryKey: ["matches"],
		queryFn: async () => {
			const response = await listMatches();
			return response.data?.items ?? [];
		}
	}));

	const leaderboardQuery = createQuery(() => ({
		queryKey: ["leaderboard"],
		queryFn: async () => {
			const response = await getLeaderboard();
			return response.data?.items ?? [];
		}
	}));
</script>

<main>
	<h1>Office Foosball</h1>
	<p>MVP dashboard scaffolded with OpenAPI-generated client and TanStack Query.</p>

	<section>
		<h2>Players</h2>
		{#if playersQuery.isLoading}
			<p>Loading players...</p>
		{:else if playersQuery.isError}
			<p>Failed to load players.</p>
		{:else if (playersQuery.data?.length ?? 0) === 0}
			<p>No players yet.</p>
		{:else}
			<ul>
				{#each playersQuery.data ?? [] as player}
					<li>{player.name}</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section>
		<h2>Matches</h2>
		{#if matchesQuery.isLoading}
			<p>Loading matches...</p>
		{:else if matchesQuery.isError}
			<p>Failed to load matches.</p>
		{:else if (matchesQuery.data?.length ?? 0) === 0}
			<p>No matches logged yet.</p>
		{:else}
			<ul>
				{#each matchesQuery.data ?? [] as match}
					<li>{match.leftScore} - {match.rightScore} ({new Date(match.playedAt).toLocaleString()})</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section>
		<h2>Leaderboard</h2>
		{#if leaderboardQuery.isLoading}
			<p>Loading standings...</p>
		{:else if leaderboardQuery.isError}
			<p>Failed to load standings.</p>
		{:else if (leaderboardQuery.data?.length ?? 0) === 0}
			<p>No leaderboard data yet.</p>
		{:else}
			<table>
				<thead>
					<tr>
						<th>Player</th>
						<th>W</th>
						<th>L</th>
						<th>GD</th>
					</tr>
				</thead>
				<tbody>
					{#each leaderboardQuery.data ?? [] as row}
						<tr>
							<td>{row.playerName}</td>
							<td>{row.wins}</td>
							<td>{row.losses}</td>
							<td>{row.goalDifference}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</section>
</main>

<style>
	main {
		margin: 0 auto;
		max-width: 54rem;
		padding: 1.5rem;
		font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
	}

	h1 {
		margin-bottom: 0.5rem;
	}

	section {
		margin-top: 2rem;
	}

	ul {
		padding-left: 1.25rem;
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	th,
	td {
		padding: 0.5rem;
		border-bottom: 1px solid #ddd;
		text-align: left;
	}
</style>
