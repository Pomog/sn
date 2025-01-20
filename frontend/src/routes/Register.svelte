<script>
    import Register from '../components/Register.svelte';

    function handleRegister(event) {
        console.log('User registered:', event);
        // Handle the user data after registration, e.g., store it or show a message
    }
</script>

<Register onRegister={handleRegister} />
