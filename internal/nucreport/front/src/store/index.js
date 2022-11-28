import {createStore} from 'vuex';
import createPersistedState from "vuex-persistedstate";

export default new createStore({
    state: {},
    mutations: {
        login(state, payload) {
            state.user = payload.user;
            state.userId = payload.userId;
        },
        logout(state) {
            state.user = null;
            state.userId = null;
        }
    },
    actions: {},
    modules: {},
    plugins: [createPersistedState()]
});
