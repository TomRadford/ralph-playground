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

<main>
	<h1>Office Foosball</h1>
	<p>Log players and matches, then track standings from those results.</p>

	<section>
		<h2>Add player</h2>
		<form on:submit={handleCreatePlayer}>
			<label for="player-name">Name</label>
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
				<button type="submit" disabled={createPlayerMutation.isPending}>Add player</button>
			</div>
		</form>
		{#if playerFormError}
			<p class="message error">{playerFormError}</p>
		{/if}
		{#if playerFormSuccess}
			<p class="message success">{playerFormSuccess}</p>
		{/if}
	</section>

	<section>
		<h2>Log match</h2>
		{#if (playersQuery.data?.length ?? 0) < 2}
			<p>Add at least two players before logging a match.</p>
		{:else}
			<form on:submit={handleCreateMatch}>
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
					<label>
						Played at
						<input type="datetime-local" bind:value={playedAtLocal} required />
					</label>
				</div>
				<button type="submit" disabled={createMatchMutation.isPending}>Log match</button>
			</form>
		{/if}
		{#if matchFormError}
			<p class="message error">{matchFormError}</p>
		{/if}
		{#if matchFormSuccess}
			<p class="message success">{matchFormSuccess}</p>
		{/if}
	</section>

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
					<li>
						{describeSide(match.left.player1Id, match.left.player2Id)} {match.leftScore}
						-
						{match.rightScore} {describeSide(match.right.player1Id, match.right.player2Id)}
						({new Date(match.playedAt).toLocaleString()})
					</li>
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

	form {
		margin-top: 0.75rem;
	}

	label {
		display: block;
		font-size: 0.95rem;
		font-weight: 600;
	}

	.form-row {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.4rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}

	input,
	select,
	button {
		border: 1px solid #ccc;
		border-radius: 6px;
		padding: 0.5rem 0.65rem;
		font: inherit;
	}

	button {
		background: #111;
		color: #fff;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.65;
		cursor: default;
	}

	.message {
		margin-top: 0.5rem;
	}

	.message.error {
		color: #a70f0f;
	}

	.message.success {
		color: #176f2c;
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
