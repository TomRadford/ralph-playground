<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import {
		createMatch,
		createPlayer,
		getLeaderboard,
		listMatches,
		listPlayers,
		type CreateMatchRequest,
		type Player
	} from "$lib/api";

	const queryClient = useQueryClient();

	let playerName = "";
	let playerFormError = "";
	let playerFormSuccess = "";

	let leftPlayer1Id = "";
	let leftPlayer2Id = "";
	let rightPlayer1Id = "";
	let rightPlayer2Id = "";
	let leftScore = 0;
	let rightScore = 0;
	let playedAtLocal = getNowLocalDateTimeValue();
	let matchFormError = "";
	let matchFormSuccess = "";

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

	const createPlayerMutation = createMutation(() => ({
		mutationFn: async (name: string) => {
			const response = await createPlayer({ body: { name } });
			if (response.error) {
				throw new Error(response.error.message);
			}
			if (!response.data) {
				throw new Error("Player creation failed.");
			}
			return response.data;
		},
		onSuccess: async () => {
			playerName = "";
			playerFormError = "";
			playerFormSuccess = "Player created.";
			await queryClient.invalidateQueries({ queryKey: ["players"] });
			await queryClient.invalidateQueries({ queryKey: ["leaderboard"] });
		}
	}));

	const createMatchMutation = createMutation(() => ({
		mutationFn: async (body: CreateMatchRequest) => {
			const response = await createMatch({ body });
			if (response.error) {
				throw new Error(response.error.message);
			}
			if (!response.data) {
				throw new Error("Match creation failed.");
			}
			return response.data;
		},
		onSuccess: async () => {
			leftPlayer1Id = "";
			leftPlayer2Id = "";
			rightPlayer1Id = "";
			rightPlayer2Id = "";
			leftScore = 0;
			rightScore = 0;
			playedAtLocal = getNowLocalDateTimeValue();
			matchFormError = "";
			matchFormSuccess = "Match logged.";
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: ["matches"] }),
				queryClient.invalidateQueries({ queryKey: ["leaderboard"] })
			]);
		}
	}));

	function getNowLocalDateTimeValue(): string {
		const now = new Date();
		now.setSeconds(0, 0);
		const localIso = new Date(now.getTime() - now.getTimezoneOffset() * 60000).toISOString();
		return localIso.slice(0, 16);
	}

	function getErrorMessage(error: unknown): string {
		if (error instanceof Error && error.message) {
			return error.message;
		}
		return "Request failed.";
	}

	function parsePlayedAt(localDateTime: string): string | null {
		const parsed = new Date(localDateTime);
		if (Number.isNaN(parsed.getTime())) {
			return null;
		}
		return parsed.toISOString();
	}

	function sideLabel(player: Player | undefined): string {
		if (!player) {
			return "Unknown player";
		}
		return player.name;
	}

	function describeSide(player1Id: string, player2Id?: string | null): string {
		const allPlayers = playersQuery.data ?? [];
		const player1 = allPlayers.find((player) => player.id === player1Id);
		const player2 = player2Id ? allPlayers.find((player) => player.id === player2Id) : undefined;

		if (!player2) {
			return sideLabel(player1);
		}

		return `${sideLabel(player1)} + ${sideLabel(player2)}`;
	}

	function formatPlayedAt(playedAt: string): string {
		return new Date(playedAt).toLocaleString();
	}

	async function handleCreatePlayer(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		playerFormError = "";
		playerFormSuccess = "";

		const name = playerName.trim();
		if (name.length < 1 || name.length > 100) {
			playerFormError = "Name must be between 1 and 100 characters.";
			return;
		}

		try {
			await createPlayerMutation.mutateAsync(name);
		} catch (error) {
			playerFormError = getErrorMessage(error);
		}
	}

	async function handleCreateMatch(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		matchFormError = "";
		matchFormSuccess = "";

		if (!leftPlayer1Id || !rightPlayer1Id) {
			matchFormError = "Both sides must include a primary player.";
			return;
		}

		if (leftScore < 0 || rightScore < 0) {
			matchFormError = "Scores must be zero or greater.";
			return;
		}

		const playedAt = parsePlayedAt(playedAtLocal);
		if (!playedAt) {
			matchFormError = "Played at must be a valid date and time.";
			return;
		}

		const playerIds = [leftPlayer1Id, leftPlayer2Id, rightPlayer1Id, rightPlayer2Id].filter(
			(playerId) => playerId !== ""
		);

		if (new Set(playerIds).size !== playerIds.length) {
			matchFormError = "A player cannot appear multiple times in one match.";
			return;
		}

		const body: CreateMatchRequest = {
			left: {
				player1Id: leftPlayer1Id,
				player2Id: leftPlayer2Id || null
			},
			right: {
				player1Id: rightPlayer1Id,
				player2Id: rightPlayer2Id || null
			},
			leftScore,
			rightScore,
			playedAt
		};

		try {
			await createMatchMutation.mutateAsync(body);
		} catch (error) {
			matchFormError = getErrorMessage(error);
		}
	}
</script>

<main class="page">
	<header class="hero">
		<div>
			<p class="eyebrow">Ralph Playground</p>
			<h1>Office Foosball</h1>
			<p class="subtitle">Log players and matches, then track standings from real results.</p>
		</div>
	</header>

	<section class="cards two-up">
		<article class="card">
			<h2>Add player</h2>
			<p class="card-help">Create someone once, then use them in every match.</p>
			<form on:submit={handleCreatePlayer} class="stack">
				<label for="player-name">Player name</label>
				<div class="form-row">
					<input
						id="player-name"
						name="player-name"
						type="text"
						maxlength="100"
						placeholder="Ada Lovelace"
						bind:value={playerName}
						required
					/>
					<button type="submit" disabled={createPlayerMutation.isPending}>
						{#if createPlayerMutation.isPending}Adding...{:else}Add player{/if}
					</button>
				</div>
			</form>
			{#if playerFormError}
				<p class="message error">{playerFormError}</p>
			{/if}
			{#if playerFormSuccess}
				<p class="message success">{playerFormSuccess}</p>
			{/if}
		</article>

		<article class="card">
			<h2>Players</h2>
			{#if playersQuery.isLoading}
				<p class="status">Loading players...</p>
			{:else if playersQuery.isError}
				<p class="status error">Failed to load players.</p>
			{:else if (playersQuery.data?.length ?? 0) === 0}
				<p class="status">No players yet.</p>
			{:else}
				<ul class="player-list">
					{#each playersQuery.data ?? [] as player}
						<li>{player.name}</li>
					{/each}
				</ul>
			{/if}
		</article>
	</section>

	<section class="card">
		<h2>Log match</h2>
		<p class="card-help">Supports both 1v1 and 2v2. Leave player 2 empty for 1v1.</p>
		{#if (playersQuery.data?.length ?? 0) < 2}
			<p class="status">Add at least two players before logging a match.</p>
		{:else}
			<form on:submit={handleCreateMatch} class="stack">
				<div class="grid">
					<label>
						Left player 1
						<select bind:value={leftPlayer1Id} required>
							<option value="">Select player</option>
							{#each playersQuery.data ?? [] as player}
								<option value={player.id}>{player.name}</option>
							{/each}
						</select>
					</label>
					<label>
						Left player 2 (optional)
						<select bind:value={leftPlayer2Id}>
							<option value="">None (1v1)</option>
							{#each playersQuery.data ?? [] as player}
								<option value={player.id}>{player.name}</option>
							{/each}
						</select>
					</label>
					<label>
						Right player 1
						<select bind:value={rightPlayer1Id} required>
							<option value="">Select player</option>
							{#each playersQuery.data ?? [] as player}
								<option value={player.id}>{player.name}</option>
							{/each}
						</select>
					</label>
					<label>
						Right player 2 (optional)
						<select bind:value={rightPlayer2Id}>
							<option value="">None (1v1)</option>
							{#each playersQuery.data ?? [] as player}
								<option value={player.id}>{player.name}</option>
							{/each}
						</select>
					</label>
					<label>
						Left score
						<input type="number" min="0" bind:value={leftScore} required />
					</label>
					<label>
						Right score
						<input type="number" min="0" bind:value={rightScore} required />
					</label>
					<label class="span-2">
						Played at
						<input type="datetime-local" bind:value={playedAtLocal} required />
					</label>
				</div>
				<button class="primary" type="submit" disabled={createMatchMutation.isPending}>
					{#if createMatchMutation.isPending}Logging...{:else}Log match{/if}
				</button>
			</form>
		{/if}
		{#if matchFormError}
			<p class="message error">{matchFormError}</p>
		{/if}
		{#if matchFormSuccess}
			<p class="message success">{matchFormSuccess}</p>
		{/if}
	</section>

	<section class="cards two-up">
		<article class="card">
			<h2>Matches</h2>
			{#if matchesQuery.isLoading}
				<p class="status">Loading matches...</p>
			{:else if matchesQuery.isError}
				<p class="status error">Failed to load matches.</p>
			{:else if (matchesQuery.data?.length ?? 0) === 0}
				<p class="status">No matches logged yet.</p>
			{:else}
				<ul class="match-list">
					{#each matchesQuery.data ?? [] as match}
						<li>
							<p class="match-teams">
								{describeSide(match.left.player1Id, match.left.player2Id)}
								<span class="score">{match.leftScore} - {match.rightScore}</span>
								{describeSide(match.right.player1Id, match.right.player2Id)}
							</p>
							<p class="match-meta">{formatPlayedAt(match.playedAt)}</p>
						</li>
					{/each}
				</ul>
			{/if}
		</article>

		<article class="card">
			<h2>Leaderboard</h2>
			{#if leaderboardQuery.isLoading}
				<p class="status">Loading standings...</p>
			{:else if leaderboardQuery.isError}
				<p class="status error">Failed to load standings.</p>
			{:else if (leaderboardQuery.data?.length ?? 0) === 0}
				<p class="status">No leaderboard data yet.</p>
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
		</article>
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		background: linear-gradient(180deg, #eef6ff 0%, #f7f9fc 45%, #ffffff 100%);
	}

	.page {
		margin: 0 auto;
		max-width: 70rem;
		padding: 1.5rem;
		font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
		color: #1f2a37;
	}

	.hero {
		padding: 1.25rem 1.5rem;
		border-radius: 16px;
		background: #0b2447;
		color: #f8fbff;
		box-shadow: 0 8px 30px rgba(11, 36, 71, 0.22);
	}

	.eyebrow {
		margin: 0;
		font-size: 0.75rem;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		opacity: 0.75;
	}

	h1 {
		margin: 0.4rem 0 0.5rem;
		font-size: clamp(1.6rem, 2.4vw, 2.2rem);
	}

	.subtitle {
		margin: 0;
		max-width: 38rem;
		opacity: 0.92;
	}

	.cards {
		margin-top: 1.1rem;
		display: grid;
		gap: 1rem;
	}

	.two-up {
		grid-template-columns: repeat(auto-fit, minmax(20rem, 1fr));
	}

	.card {
		margin-top: 1rem;
		padding: 1rem 1.1rem;
		border-radius: 14px;
		background: #ffffff;
		border: 1px solid #dfe6ef;
		box-shadow: 0 6px 18px rgba(14, 32, 58, 0.06);
	}

	.card h2 {
		margin: 0 0 0.25rem;
		font-size: 1.12rem;
	}

	.card-help {
		margin: 0 0 0.85rem;
		color: #5b6777;
		font-size: 0.92rem;
	}

	.stack {
		display: grid;
		gap: 0.75rem;
	}

	label {
		display: grid;
		gap: 0.35rem;
		font-size: 0.9rem;
		font-weight: 600;
	}

	.form-row {
		display: flex;
		gap: 0.6rem;
	}

	.form-row input {
		flex: 1;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
		gap: 0.8rem;
	}

	.span-2 {
		grid-column: span 2;
	}

	@media (max-width: 38rem) {
		.span-2 {
			grid-column: span 1;
		}
	}

	input,
	select,
	button {
		border: 1px solid #c7d4e2;
		border-radius: 10px;
		padding: 0.58rem 0.7rem;
		font: inherit;
		background: #fff;
		color: #1f2a37;
	}

	input:focus,
	select:focus {
		outline: 2px solid #74a9ff;
		outline-offset: 1px;
	}

	button {
		background: #194b9a;
		border-color: #194b9a;
		color: #fff;
		font-weight: 600;
		cursor: pointer;
		transition: transform 0.05s ease, opacity 0.2s ease;
	}

	button:hover:enabled {
		opacity: 0.95;
	}

	button:active:enabled {
		transform: translateY(1px);
	}

	button:disabled {
		opacity: 0.62;
		cursor: default;
	}

	.primary {
		justify-self: start;
		padding-inline: 1rem;
	}

	.status {
		margin: 0.65rem 0 0;
		padding: 0.62rem 0.72rem;
		border-radius: 9px;
		background: #f5f8fc;
		color: #44556b;
	}

	.status.error {
		background: #fff0f0;
		color: #9b1f1f;
	}

	.message {
		margin: 0.75rem 0 0;
		padding: 0.6rem 0.7rem;
		border-radius: 9px;
		font-weight: 500;
	}

	.message.error {
		background: #fff0f0;
		color: #9b1f1f;
	}

	.message.success {
		background: #eefbf1;
		color: #196a30;
	}

	.player-list,
	.match-list {
		margin: 0.65rem 0 0;
		padding: 0;
		list-style: none;
	}

	.player-list li,
	.match-list li {
		padding: 0.65rem 0.2rem;
		border-bottom: 1px solid #e8edf3;
	}

	.player-list li:last-child,
	.match-list li:last-child {
		border-bottom: 0;
		padding-bottom: 0.1rem;
	}

	.match-teams {
		margin: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
	}

	.score {
		padding: 0.15rem 0.45rem;
		border-radius: 999px;
		background: #edf3ff;
		color: #194b9a;
		font-weight: 700;
	}

	.match-meta {
		margin: 0.24rem 0 0;
		color: #637286;
		font-size: 0.86rem;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		margin-top: 0.6rem;
	}

	th,
	td {
		padding: 0.55rem;
		border-bottom: 1px solid #e7edf4;
		text-align: left;
	}

	th {
		font-size: 0.82rem;
		color: #4d617a;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
</style>
