import {createApp} from "vue";
import App from "./App.vue";
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import {pinia} from "./store/useStore.ts";
import {router} from "./router/useRouter.ts"
import naive from 'naive-ui'

const app = createApp(App);
app.use(ElementPlus)
app.use(naive)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}
app.use(pinia)
app.use(router)

app.mount("#app");
