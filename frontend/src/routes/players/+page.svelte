<script lang="ts">
  import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
  import { createPlayer, deletePlayer, listPlayers } from "$lib/api";

  const queryClient = useQueryClient();

  let playerName = $state("");
  let playerFormError = $state("");
  let playerFormSuccess = $state("");

  const playersQuery = createQuery(() => ({
    queryKey: ["players"],
    queryFn: async () => {
      const response = await listPlayers();
      return response.data?.items ?? [];
    },
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
    },
  }));

  const deletePlayerMutation = createMutation(() => ({
    mutationFn: async (playerId: string) => {
      const response = await deletePlayer({ path: { playerId } });
      if (response.error) {
        throw new Error(response.error.message);
      }
    },
    onSuccess: async () => {
      playerFormError = "";
      playerFormSuccess = "Player deleted.";
      await queryClient.invalidateQueries({ queryKey: ["players"] });
      await queryClient.invalidateQueries({ queryKey: ["leaderboard"] });
    },
  }));

  function getErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message) {
      return error.message;
    }
    return "Request failed.";
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

  async function handleDeletePlayer(playerId: string, name: string): Promise<void> {
    playerFormError = "";
    playerFormSuccess = "";

    const shouldDelete = confirm(`Delete player "${name}"?`);
    if (!shouldDelete) {
      return;
    }

    try {
      await deletePlayerMutation.mutateAsync(playerId);
    } catch (error) {
      playerFormError = getErrorMessage(error);
    }
  }
</script>

<section class="card">
  <h2>Players</h2>
  <p class="card-help">Create players once, then reuse them in all matches.</p>

  <form onsubmit={handleCreatePlayer} class="stack">
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
</section>

<section class="card">
  <h2>Current players</h2>
  {#if playersQuery.isLoading}
    <p class="status">Loading players...</p>
  {:else if playersQuery.isError}
    <p class="status error">Failed to load players.</p>
  {:else if (playersQuery.data?.length ?? 0) === 0}
    <p class="status">No players yet.</p>
  {:else}
    <ul class="player-list">
      {#each playersQuery.data ?? [] as player}
        <li class="list-item">
          <span>{player.name}</span>
          <button
            type="button"
            class="danger"
            disabled={deletePlayerMutation.isPending}
            onclick={() => handleDeletePlayer(player.id, player.name)}
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

  .form-row {
    display: flex;
    gap: 0.6rem;
  }

  .form-row input {
    flex: 1;
  }

  input,
  button {
    border: 1px solid #c7d4e2;
    border-radius: 10px;
    padding: 0.58rem 0.7rem;
    font: inherit;
    background: #fff;
    color: #1f2a37;
  }

  input:focus {
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

  .player-list {
    margin: 0.65rem 0 0;
    padding: 0;
    list-style: none;
  }

  .player-list li {
    padding: 0.65rem 0.2rem;
    border-bottom: 1px solid #e8edf3;
  }

  .list-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .player-list li:last-child {
    border-bottom: 0;
    padding-bottom: 0.1rem;
  }
</style>
