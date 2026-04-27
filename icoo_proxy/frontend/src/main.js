import { createApp } from "vue";
import { createPinia } from "pinia";

import App from "./App.vue";
import router from "./router";
import { useUiPrefsStore } from "./stores/uiPrefs";
import "./main.css";

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

const uiPrefs = useUiPrefsStore();
uiPrefs.init();

app.mount("#app");
