import {createApp} from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import SuiVue from 'semantic-ui-vue';
import Toaster from '@meforma/vue-toaster';
// import VueClipboard from 'vue-clipboard2'
import {apiUrl} from '@/config';
import axios from 'axios';

axios.defaults.baseURL = apiUrl;
axios.defaults.withCredentials = true;

store.$http = axios;

let app = createApp(App)
    .use(Toaster)
    .use(router)
    .use(store)
    .use(SuiVue);

app.config.globalProperties.$http = axios;

app.mount("#app");