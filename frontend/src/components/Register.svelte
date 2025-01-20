<script>
    export let onRegister;

    let email = '';
    let password = '';
    let firstName = '';
    let lastName = '';
    let nickname = '';
    let image = '';
    let about = '';
    let birthday = '';
    let privateAccount = false;

    // Handle form submission
    async function handleSubmit(event) {
        event.preventDefault();

        const userData = {
            email,
            password,
            firstName,
            lastName,
            nickname,
            image,
            about,
            birthday: birthday ? new Date(birthday) : null,
            private: privateAccount,
        };

        try {
            const response = await fetch('http://localhost:8080/api/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userData),
            });

            if (!response.ok) {
                throw new Error('Registration failed');
            }

            const user = await response.json();
            if (onRegister) {
                onRegister(user);  // Call the parent callback with the user data
            }
            alert('Registration successful!');
        } catch (error) {
            console.error('Error during registration:', error);
            alert('An error occurred during registration.');
        }
    }
</script>

<h2>Register</h2>
<form on:submit={handleSubmit}>
    <label for="email">Email</label>
    <input type="email" id="email" bind:value={email} required />

    <label for="password">Password</label>
    <input type="password" id="password" bind:value={password} required />

    <label for="firstName">First Name</label>
    <input type="text" id="firstName" bind:value={firstName} required />

    <label for="lastName">Last Name</label>
    <input type="text" id="lastName" bind:value={lastName} required />

    <label for="nickname">Nickname</label>
    <input type="text" id="nickname" bind:value={nickname} />

    <label for="image">Profile Image URL</label>
    <input type="text" id="image" bind:value={image} />

    <label for="about">About You</label>
    <textarea id="about" bind:value={about}></textarea>

    <label for="birthday">Birthday</label>
    <input type="date" id="birthday" bind:value={birthday} />

    <label for="private">Private Account</label>
    <input type="checkbox" id="private" bind:checked={privateAccount} />

    <button type="submit">Register</button>
</form>
