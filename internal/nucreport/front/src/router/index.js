import IndexPage from '@/views/IndexPage';
import RegisterPage from "@/views/RegisterPage";
import LoginPage from "@/views/LoginPage";
import HomePage from "@/views/HomePage";

import { createRouter,createWebHistory} from 'vue-router'

const routes = [
    {
        path: '/',
        name: 'Index',
        component: IndexPage
    },
    {
        path: '/register',
        name: 'Register',
        component: RegisterPage,
    },
    {
        path: '/login',
        name: 'Login',
        component: LoginPage,
    },
    {
        path: '/home',
        name: 'Home',
        component: HomePage,
    },
];

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});

export default router;