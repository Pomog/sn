<script>
    import { session } from "../stores.js";
    import axios from "axios";

    let email = "";
    let password = "";

    async function login() {
        try {
            const response = await axios.post("http://localhost:8080/login", { email, password });
            session.set({ token: response.data.token, user: response.data.user });
            alert("Login successful!");
        } catch (err) {
            console.error("Login failed:", err);
            alert("Login failed. Check your credentials.");
        }
    }
</script>

<form on:submit|preventDefault={login}>
    <label for="email">Email:</label>
    <input id="email" type="email" bind:value={email} />

    <label for="password">Password:</label>
    <input id="password" type="password" bind:value={password} />

    <button type="submit">Login</button>
</form>
