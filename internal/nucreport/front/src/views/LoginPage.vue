<template>
  <page-layout>
    <div class="ui text container">
      <div class="ui one column grid">
        <div class="column">
          <h1 class="ui header">Login</h1>
        </div>
      </div>
      <div class="ui text container">
        <form @submit="login" class="ui form" method="post" action="">
          <div class="field">
            <label>username</label>
            <input type="text" required name="username" v-model="username" placeholder="user">
          </div>
          <div class="field">
            <label>Password</label>
            <input type="password" id="password" required name="password" v-model="password" placeholder="">
          </div>
          <button class="ui button" type="submit">Submit</button>
        </form>
      </div>
    </div>
  </page-layout>
</template>
<script>
import PageLayout from './PageLayout';

export default {
  components: {
    PageLayout
  },
  methods: {
    async login(event) {
      event.preventDefault();
      try {
        let resp = await this.$http.post('login', {
          username: this.username,
          password: this.password,
        });

        this.$store.commit('login', {user: resp.data.username, userId: resp.data.id});
        this.$router.push({name: 'Home'}).catch(() => {
        });
      } catch (error) {
        this.$toast.error(JSON.stringify(error.response.data.error), {position: 'top-right'});
      }
    },
  },
  data: function () {
    return {
      username: null,
      password: null,
      error: null,
      errorVisible: true,
    };
  },
};
</script>