<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";
  import { getLeaderboard } from "$lib/api";

  const leaderboardQuery = createQuery(() => ({
    queryKey: ["leaderboard"],
    queryFn: async () => {
      const response = await getLeaderboard();
      return response.data?.items ?? [];
    },
  }));
</script>

<section class="card">
  <h2>Leaderboard</h2>
  <p class="card-help">Rankings are generated directly from logged match results.</p>

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
