<template>
  <page-layout>
    <div class="ui text container">
      <p>Welcome, {{ user }}</p>

      <div class="ui message">
        <div class="header">
          Report upload
        </div>
        <p>Use SFTP to upload your report files</p>
        <p><code>sftp -P 4222 {{ user }}@<span id="host">{{ host }}</span></code></p>
        <p>The password is the same as your account password. </p>
      </div>

      <div class="ui message">
        Uploaded new report files ? Reindex your directory to see them!
        <button type="submit" @click="reindex">Reindex</button>
      </div>
    </div>

    <br>
    <div class="ui text container">
      <p>Your files</p>
      <div class="ui list" v-for="file in indexedFiles" :key="file.path">
        <div class="item">
          <a :href="'/api/file?path=' + file.path + '&token='">{{ file.path }}</a> |||
          <a :href="'/api/file?path=' + file.path + '&token=' + file.token">Share with other engineer!</a>
        </div>
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
  data() {
    return {
      indexedFiles: [],
    }
  },
  computed: {
    user() {
      return this.$store.state.user;
    },
    host() {
      return window.location.hostname;
    }
  },
  async mounted() {
    await this.getIndexedFiles();
  },
  methods: {
    async getIndexedFiles() {
      try {
        let resp = await this.$http.get('files');
        this.indexedFiles = resp.data;
      } catch (error) {
        this.$toast.error(JSON.stringify(error.response.data.error), {position: 'top-right'});
      }
    },
    async reindex() {
      try {
        await this.$http.post('reindex');
        await this.getIndexedFiles();
      } catch (error) {
        this.$toast.error(JSON.stringify(error.response.data.error), {position: 'top-right'});
      }
    }
  }
};
</script>
