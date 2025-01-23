<template>
  <div class="flex w-screen h-screen overflow-hidden">
    <div class="flex items-center justify-center px-6 mx-auto w-full lg:w-1/2">
      <div class="flex flex-col lg:w-4/6 w-full min-w-[250px]">
        <div class="flex w-full items-start flex-col">
          <div class="flex justify-center mx-auto">
            <NuxtLink to="/">
              <img class="w-auto h-10 mb-4" src="assets/images/logo-icon.png" alt="Logo">
            </NuxtLink>
          </div>

          <p class="mt-3 font-bold text-4xl mb-4 text-gray-800">Sign in</p>
          <p class="text-gray-600">Welcome back, please enter your credentials!</p>
        </div>

        <div class="mt-8 w-full flex items-center justify-center flex-col">
          <UForm ref="formEl" :schema="schema" :state="state" class="space-y-5 w-full" @submit="onSubmit">
            <UFormGroup label="Email" name="email">
              <UInput size="lg" v-model="state.email" placeholder="Enter your email" class="border-gray-300" />
            </UFormGroup>

            <UFormGroup label="Password" name="password">
              <UInput size="lg" v-model="state.password" type="password" placeholder="•••••••••" autocomplete="on" class="border-gray-300" />
            </UFormGroup>

            <div class="flex w-full justify-between">
              <UCheckbox v-model="state.remember" label="Remember me" name="remember" />
              <ULink class="font-bold text-sm text-primary">Forgot your password?</ULink>
            </div>

            <UButton type="submit" size="lg" class="button bg-indigo-600 text-white w-full cursor-pointer hover:bg-indigo-700 transition-all" block>
              Sign in
            </UButton>
          </UForm>
          <p class="mt-8 text-md text-center text-gray-500">Don't have an account yet? <ULink to="/register" class="text-indigo-600 hover:text-indigo-800">Register</ULink>.</p>
        </div>
      </div>
    </div>

    <div class="bg-cover hidden lg:block lg:w-1/2 bg-gradient-to-r from-teal-800 to-purple-800">
      <div class="flex items-center h-full px-20 bg-black bg-opacity-60">
        <div class="flex w-full flex-col items-center justify-center">
          <div class="h-70 w-full max-w-[500px] bg-white/30 mb-20 rounded-xl"></div>

          <!-- Heading -->
          <h2 class="text-4xl font-extrabold text-white sm:text-3xl text-center tracking-tight leading-snug">
            Welcome to Your Social Network
          </h2>

          <!-- Description -->
          <p class="max-w-xl mt-4 text-white text-center leading-relaxed">
            This is a learning project from Kood/Jõhvi, where we are building a Facebook-like social network. The app will include features such as:
          </p>

          <!-- Features List -->
          <ul class="list-disc text-white mt-6 space-y-3">
            <li><strong class="text-teal-300">Followers</strong> – Follow your friends and stay updated on their activities.</li>
            <li><strong class="text-teal-300">Profile</strong> – Create and personalize your own profile page.</li>
            <li><strong class="text-teal-300">Posts</strong> – Share updates, photos, and thoughts with your network.</li>
            <li><strong class="text-teal-300">Groups</strong> – Join groups of interest and engage in community discussions.</li>
            <li><strong class="text-teal-300">Notifications</strong> – Get real-time notifications about activities, messages, and more.</li>
            <li><strong class="text-teal-300">Chats</strong> – Communicate with your friends and groups through private messages.</li>
          </ul>

          <!-- Footer Text -->
          <p class="mt-6 text-md text-center text-gray-300">
            Stay tuned for new features as we continue developing the application!
          </p>
        </div>
      </div>
    </div>


  </div>
</template>

<script setup lang="ts">
import { z } from 'zod'
import type { FormSubmitEvent } from '#ui/types'

const { login } = useAuth();
const formEl = ref();

useHead({
  title: "Sign in",
})

definePageMeta({
  alias: ["/login"],
  layout: "auth",
  middleware: ["guest-only"],
});

const schema = z.object({
  email: z.string().email('Invalid email'),
  remember: z.boolean().optional(),
  password: z.string().min(8, 'Must be at least 8 characters')
})

type Schema = z.output<typeof schema>

const state = reactive({
  email: undefined,
  remember: undefined,
  password: undefined
})

async function onSubmit(event: FormSubmitEvent<Schema>) {
  try {
    await login(event.data.email, event.data.password, event.data.remember ? true : false)
    const redirect = '/'
    await navigateTo(redirect);
  } catch (error) {
    console.log(error);

    formEl.value.setErrors([
      {
        message: "Invalid email or password",
        path: "email",
      }, {
        message: "Invalid email or password",
        path: "",
      }
    ])
  }
}
</script>