<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";
  import { getLeaderboard, listMatches, listPlayers } from "$lib/api";

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

  const leaderboardQuery = createQuery(() => ({
    queryKey: ["leaderboard"],
    queryFn: async () => {
      const response = await getLeaderboard();
      return response.data?.items ?? [];
    },
  }));
</script>

<section class="overview card">
  <h2>Overview</h2>
  <p>Use the sections below to manage players, log matches, and track standings.</p>
</section>

<section class="cards">
  <a href="/players" class="card link-card">
    <h3>Players</h3>
    <p>Create and remove players for your office table.</p>
    <strong>{playersQuery.data?.length ?? 0} total players</strong>
  </a>
  <a href="/matches" class="card link-card">
    <h3>Matches</h3>
    <p>Log new results and clean up incorrect entries.</p>
    <strong>{matchesQuery.data?.length ?? 0} matches logged</strong>
  </a>
  <a href="/leaderboard" class="card link-card">
    <h3>Leaderboard</h3>
    <p>See who is leading by wins, losses, and goal difference.</p>
    <strong>{leaderboardQuery.data?.length ?? 0} ranked players</strong>
  </a>
</section>

<style>
  .overview {
    margin-top: 0.2rem;
  }

  .overview p {
    margin: 0.4rem 0 0;
    color: #516073;
  }

  .cards {
    margin-top: 0.9rem;
    display: grid;
    gap: 0.9rem;
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  }

  .card {
    padding: 1rem 1.1rem;
    border-radius: 14px;
    background: #ffffff;
    border: 1px solid #dfe6ef;
    box-shadow: 0 6px 18px rgba(14, 32, 58, 0.06);
  }

  .link-card {
    text-decoration: none;
    color: inherit;
    transition: transform 0.08s ease;
  }

  .link-card:hover {
    transform: translateY(-1px);
  }

  h2,
  h3 {
    margin: 0;
  }

  h3 {
    margin-bottom: 0.45rem;
  }

  .link-card p {
    margin: 0;
    color: #5b6777;
  }

  .link-card strong {
    margin-top: 0.75rem;
    display: block;
    color: #194b9a;
  }
</style>
