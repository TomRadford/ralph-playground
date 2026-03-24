<script lang="ts">
  import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
  import { createMatch, deleteMatch, listMatches, listPlayers, type CreateMatchRequest, type Player } from "$lib/api";

  const queryClient = useQueryClient();

  let leftPlayer1Id = $state("");
  let leftPlayer2Id = $state("");
  let rightPlayer1Id = $state("");
  let rightPlayer2Id = $state("");
  let leftScore = $state(0);
  let rightScore = $state(0);
  let playedAtLocal = $state(getNowLocalDateTimeValue());
  let matchFormError = $state("");
  let matchFormSuccess = $state("");

  const playersQuery = createQuery(() => ({
    queryKey: ["players"],
    queryFn: async () => {
      const response = await listPlayers();
      return response.data?.items ?? [];
    },
  }));

  const matchesQuery = createQuery(() => ({
    queryKey: ["matches"],
    queryFn: async () => {
      const response = await listMatches();
      return response.data?.items ?? [];
    },
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
        queryClient.invalidateQueries({ queryKey: ["leaderboard"] }),
      ]);
    },
  }));

  const deleteMatchMutation = createMutation(() => ({
    mutationFn: async (matchId: string) => {
      const response = await deleteMatch({ path: { matchId } });
      if (response.error) {
        throw new Error(response.error.message);
      }
    },
    onSuccess: async () => {
      matchFormError = "";
      matchFormSuccess = "Match deleted.";
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["matches"] }),
        queryClient.invalidateQueries({ queryKey: ["leaderboard"] }),
      ]);
    },
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
      (playerId) => playerId !== "",
    );
    if (new Set(playerIds).size !== playerIds.length) {
      matchFormError = "A player cannot appear multiple times in one match.";
      return;
    }

    const body: CreateMatchRequest = {
      left: {
        player1Id: leftPlayer1Id,
        player2Id: leftPlayer2Id || null,
      },
      right: {
        player1Id: rightPlayer1Id,
        player2Id: rightPlayer2Id || null,
      },
      leftScore,
      rightScore,
      playedAt,
    };

    try {
      await createMatchMutation.mutateAsync(body);
    } catch (error) {
      matchFormError = getErrorMessage(error);
    }
  }

  async function handleDeleteMatch(matchId: string): Promise<void> {
    matchFormError = "";
    matchFormSuccess = "";

    const shouldDelete = confirm("Delete this match result?");
    if (!shouldDelete) {
      return;
    }

    try {
      await deleteMatchMutation.mutateAsync(matchId);
    } catch (error) {
      matchFormError = getErrorMessage(error);
    }
  }
</script>

<section class="card">
  <h2>Log match</h2>
  <p class="card-help">Supports both 1v1 and 2v2. Leave player 2 empty for 1v1.</p>
  {#if (playersQuery.data?.length ?? 0) < 2}
    <p class="status">Add at least two players before logging a match.</p>
  {:else}
    <form onsubmit={handleCreateMatch} class="stack">
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

<section class="card">
  <h2>Recent matches</h2>
  {#if matchesQuery.isLoading}
    <p class="status">Loading matches...</p>
  {:else if matchesQuery.isError}
    <p class="status error">Failed to load matches.</p>
  {:else if (matchesQuery.data?.length ?? 0) === 0}
    <p class="status">No matches logged yet.</p>
  {:else}
    <ul class="match-list">
      {#each matchesQuery.data ?? [] as match}
        <li class="list-item">
          <div>
            <p class="match-teams">
              {describeSide(match.left.player1Id, match.left.player2Id)}
              <span class="score">{match.leftScore} - {match.rightScore}</span>
              {describeSide(match.right.player1Id, match.right.player2Id)}
            </p>
            <p class="match-meta">{formatPlayedAt(match.playedAt)}</p>
          </div>
          <button
            type="button"
            class="danger"
            disabled={deleteMatchMutation.isPending}
            onclick={() => handleDeleteMatch(match.id)}
          >
            Delete
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .card {
    margin-top: 0.85rem;
    padding: 1rem 1.1rem;
    border-radius: 14px;
    background: #ffffff;
    border: 1px solid #dfe6ef;
    box-shadow: 0 6px 18px rgba(14, 32, 58, 0.06);
  }

  h2 {
    margin: 0;
    font-size: 1.1rem;
  }

  .card-help {
    margin: 0.35rem 0 0.85rem;
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
  }

  button:disabled {
    opacity: 0.62;
    cursor: default;
  }

  .danger {
    background: #b83939;
    border-color: #b83939;
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

  .match-list {
    margin: 0.65rem 0 0;
    padding: 0;
    list-style: none;
  }

  .match-list li {
    padding: 0.65rem 0.2rem;
    border-bottom: 1px solid #e8edf3;
  }

  .list-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

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
</style>
