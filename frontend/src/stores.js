import { writable } from "svelte/store";

// Store for user session data
export const session = writable({
    token: null,
    user: null,
});
