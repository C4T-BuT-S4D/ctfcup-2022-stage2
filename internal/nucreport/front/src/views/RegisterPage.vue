<template>
  <page-layout>
    <div class="ui text container">
      <div class="ui one column grid">
        <div class="column">
          <h1 class="ui header">Register</h1>
        </div>
      </div>
      <div class="ui text container">
        <form @submit="register" class="ui form" method="post" action="">
          <div class="field">
            <label>Username</label>
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
    PageLayout,
  },
  methods: {
    async register(event) {
      event.preventDefault();
      try {
        await this.$http.post('register', {
          username: this.username,
          password: this.password,
        });
        this.$router.push({name: 'Login'}).catch(() => {
        });
      } catch (error) {
        this.$toast.error(JSON.stringify(error.response.data.error),  {position: 'top-right'});
      }
    },
  },
  data: function () {
    return {
      username: null,
      password: null,
    };
  },
};
</script>