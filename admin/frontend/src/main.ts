import ElementPlus from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import { createPinia } from 'pinia';
import { createApp } from 'vue';
import 'element-plus/dist/index.css';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';

import App from './App.vue';
import { permissionDirective } from './directives/permission';
import router from './router';
import './styles/index.scss';

const app = createApp(App);

// Element Plus icons registered globally in three casings so meta strings
// (e.g. "dashboard", "user") and direct component names (e.g. "Dashboard")
// both work.
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component as any);
  const kebab = key
    .replace(/([A-Z])/g, '-$1')
    .toLowerCase()
    .replace(/^-/, '');
  app.component(kebab, component as any);
  const lower = key.charAt(0).toLowerCase() + key.slice(1);
  app.component(lower, component as any);
}

app.use(createPinia());
app.use(router);
app.use(ElementPlus, { locale: zhCn });
app.directive('permission', permissionDirective);

app.mount('#app');
