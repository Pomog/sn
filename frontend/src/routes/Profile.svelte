<script>
    import { session } from "../stores.js";
</script>

{#if $session.user}
    <h1>Welcome, {$session.user.firstname}!</h1>
    <p>Email: {$session.user.email}</p>
    <p>About: {$session.user.about}</p>
{:else}
    <p>You are not logged in. Please log in to view your profile.</p>
{/if}
