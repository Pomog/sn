<script>
  import { Router, Route } from "svelte-routing";
  import Home from "./routes/Home.svelte";
  import Login from "./routes/Login.svelte";
  import Profile from "./routes/Profile.svelte";
  import Posts from "./routes/Posts.svelte";
  import Groups from "./routes/Groups.svelte";
  import Navbar from "./components/Navbar.svelte";
  import Register from "./routes/Register.svelte";
</script>

<Router>
  <Navbar />
  <Route path="/" component={Home} />
  <Route path="/login" component={Login} />
  <Route path="/profile" component={Profile} />
  <Route path="/posts" component={Posts} />
  <Route path="/groups" component={Groups} />
  <Route path="/register" component={Register} />
</Router>
