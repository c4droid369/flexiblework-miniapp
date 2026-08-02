<template>
  <div class="showcase">
    <div class="hero">
      <h1 class="title">🎓 校园灵活用工小程序模板</h1>
      <p class="subtitle">兼职招聘场景 · 学生 / 雇主 / 校园代理 · 后端 + 管理端 + 微信小程序</p>
      <div class="badges">
        <span class="badge">Go + Gin + GORM</span>
        <span class="badge">Vue 3 + Element Plus</span>
        <span class="badge">uni-app · Vue 3 · Pinia</span>
        <span class="badge">MySQL · Docker Compose</span>
      </div>
    </div>

    <div class="stage">
      <!-- Phone mockup -->
      <div class="phone-wrap">
        <div class="phone">
          <div class="notch"></div>
          <div class="screen">
            <iframe
              v-if="hasApp"
              src="/showcase-app/index.html"
              title="校园灵活用工小程序"
              @load="onLoaded"
              @error="onError"
            ></iframe>
            <div v-else class="placeholder">
              <div class="logo">🎓</div>
              <h2>校园灵活用工</h2>
              <p class="muted">下方三条体验账号任选其一登录</p>
              <div class="demo-list">
                <button class="demo-btn" @click="simulate('demo_student')">
                  <text class="role">学生</text>
                  <text class="name">demo_student / demo123</text>
                </button>
                <button class="demo-btn" @click="simulate('demo_employer')">
                  <text class="role">雇主</text>
                  <text class="name">demo_employer / demo123</text>
                </button>
                <button class="demo-btn" @click="simulate('demo_agent')">
                  <text class="role">校园代理</text>
                  <text class="name">demo_agent / demo123</text>
                </button>
              </div>
              <p class="hint">▲ 在 HBuilderX 打开 uniapp/,运行 → 发行 → 网站-PC Web 或手机H5,把输出目录 build/h5(或 unpackage/dist/dev/mp-weixin 等)复制到 admin/frontend/public/showcase-app/,iframe 会自动加载</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right side: feature highlights -->
      <div class="features">
        <h3>核心特性</h3>
        <div class="feature-item">
          <span class="icon">🔐</span>
          <div>
            <h4>三端独立认证</h4>
            <p>学生/雇主/代理三种角色,各自独立的 profile + 实名/资质审核流,JWT + RBAC perm code 中间件 O(1) 校验。</p>
          </div>
        </div>
        <div class="feature-item">
          <span class="icon">📋</span>
          <div>
            <h4>完整业务状态机</h4>
            <p>岗位(6 态) · 报名(5 态) · 订单(7 态 + Mock 支付) · 评价 · 消息推送,全程闭环可走通。</p>
          </div>
        </div>
        <div class="feature-item">
          <span class="icon">🛡️</span>
          <div>
            <h4>管理端开箱即用</h4>
            <p>分类 / 岗位审核 / 学生认证 / 雇主资质 / 校园代理资质 / 订单监控 / 评价 / 消息广播,Vue3 + Element Plus 6 个业务页面。</p>
          </div>
        </div>
        <div class="feature-item">
          <span class="icon">🐳</span>
          <div>
            <h4>一键 Docker 部署</h4>
            <p>docker compose up -d 同时拉起 MySQL + 后端 + Nginx 前端,首启自动 migration + seed + 示例数据。</p>
          </div>
        </div>

        <h3 style="margin-top: 32px">后端 API</h3>
        <ul class="endpoints">
          <li><code>GET  /api/v1/categories</code> 公开分类列表</li>
          <li><code>GET  /api/v1/jobs</code> 公开岗位列表(分页/筛选/搜索)</li>
          <li><code>POST /api/v1/auth/register</code> 注册(student/employer/agent)</li>
          <li><code>POST /api/v1/jobs/:id/apply</code> 学生报名</li>
          <li><code>POST /api/v1/employer/applications/:id/audit</code> 雇主审核</li>
          <li><code>POST /api/v1/employer/applications/:id/hire</code> 录用并下单</li>
          <li><code>POST /api/v1/orders/:id/pay|checkin|complete</code> 学生状态机</li>
          <li><code>POST /api/v1/employer/orders/:id/confirm</code> 雇主结算</li>
        </ul>
        <p class="muted hint">完整 Swagger: <code>http://localhost:8080/swagger/index.html</code></p>
      </div>
    </div>

    <div class="footer muted">
      <p>校园灵活用工小程序模板 · MIT · 后端 + 管理端 + 小程序</p>
      <p class="footer-tiny">部署指南 / API 文档 / 二次开发说明见根目录 README.md</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const hasApp = ref(false)

onMounted(async () => {
  // Probe whether the H5 build exists at /showcase-app/index.html.
  try {
    const res = await fetch('/showcase-app/index.html', { method: 'HEAD' })
    hasApp.value = res.ok
  } catch {
    hasApp.value = false
  }
})

function onLoaded() { /* iframe ready */ }
function onError() { hasApp.value = false }

function simulate(account: string) {
  // When H5 build isn't present, this hook is reserved for future demo.
  // Currently it just shows an alert pointing to the build step.
  alert(`完整 demo 需要先用 HBuilderX 构建 uniapp/ 到 H5,然后把输出放到 admin/frontend/public/showcase-app/。
\n当前账号:${account} / demo123`)
}
</script>

<style scoped>
.showcase {
  min-height: 100vh;
  background: linear-gradient(135deg, #FF7A45 0%, #FFC53D 50%, #4ECB73 100%);
  padding: 40px 24px;
  box-sizing: border-box;
}
.hero {
  text-align: center;
  color: #fff;
  margin-bottom: 40px;
}
.title {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 12px;
  text-shadow: 0 2px 12px rgba(0,0,0,0.15);
}
.subtitle {
  font-size: 16px;
  margin: 0 0 20px;
  opacity: 0.95;
}
.badges {
  display: flex;
  gap: 8px;
  justify-content: center;
  flex-wrap: wrap;
}
.badge {
  background: rgba(255,255,255,0.22);
  backdrop-filter: blur(8px);
  color: #fff;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  border: 1px solid rgba(255,255,255,0.3);
}

.stage {
  display: flex;
  gap: 60px;
  align-items: flex-start;
  justify-content: center;
  max-width: 1280px;
  margin: 0 auto;
  flex-wrap: wrap;
}

/* Phone mockup */
.phone-wrap {
  perspective: 1000px;
  flex-shrink: 0;
}
.phone {
  width: 375px;
  height: 760px;
  background: #1a1a1a;
  border-radius: 48px;
  padding: 12px;
  box-shadow: 0 30px 60px rgba(0,0,0,0.25), 0 0 0 2px rgba(255,255,255,0.08);
  position: relative;
}
.notch {
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  width: 150px;
  height: 28px;
  background: #1a1a1a;
  border-radius: 0 0 16px 16px;
  z-index: 2;
}
.screen {
  width: 100%;
  height: 100%;
  background: #fff;
  border-radius: 36px;
  overflow: hidden;
  position: relative;
}
.screen iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}

/* Placeholder when H5 build missing */
.placeholder {
  height: 100%;
  background: linear-gradient(135deg, #FFF3EB 0%, #EAF9F0 100%);
  padding: 32px 24px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow-y: auto;
}
.placeholder .logo {
  font-size: 80px;
  margin: 40px 0 16px;
}
.placeholder h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: #1F2329;
}
.placeholder .muted {
  color: #6B7785;
  font-size: 13px;
  text-align: center;
  margin: 0 0 24px;
}
.demo-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}
.demo-btn {
  width: 100%;
  background: #fff;
  border: 1px solid #E5E6EB;
  border-radius: 12px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.15s;
}
.demo-btn:hover {
  border-color: #FF7A45;
  background: #FFF3EB;
}
.demo-btn .role {
  background: linear-gradient(135deg, #FF7A45, #E5612A);
  color: #fff;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 12px;
}
.demo-btn .name {
  color: #4E5969;
  font-family: monospace;
  font-size: 13px;
}
.placeholder .hint {
  font-size: 11px;
  color: #86909C;
  background: #FFF8E1;
  border-left: 3px solid #FFC53D;
  padding: 12px;
  border-radius: 4px;
  line-height: 1.6;
  margin-top: 16px;
}

/* Right features */
.features {
  flex: 1;
  min-width: 360px;
  max-width: 540px;
  background: rgba(255,255,255,0.95);
  backdrop-filter: blur(12px);
  border-radius: 24px;
  padding: 32px;
  color: #1F2329;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1);
}
.features h3 {
  margin: 0 0 16px;
  font-size: 20px;
  color: #FF7A45;
}
.features h3:not(:first-child) {
  margin-top: 28px;
}
.feature-item {
  display: flex;
  gap: 14px;
  margin-bottom: 18px;
}
.feature-item .icon {
  font-size: 28px;
  flex-shrink: 0;
  width: 40px;
  text-align: center;
}
.feature-item h4 {
  margin: 0 0 4px;
  font-size: 15px;
  color: #1F2329;
}
.feature-item p {
  margin: 0;
  font-size: 13px;
  color: #6B7785;
  line-height: 1.6;
}
.endpoints {
  list-style: none;
  padding: 0;
  margin: 0;
}
.endpoints li {
  padding: 6px 0;
  font-size: 12px;
  color: #4E5969;
  border-bottom: 1px dashed #E5E6EB;
}
.endpoints li:last-child { border-bottom: none; }
.endpoints code {
  background: #F7F8FA;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  color: #FF7A45;
  margin-right: 8px;
}

.footer {
  text-align: center;
  margin-top: 60px;
  color: rgba(255,255,255,0.85);
}
.footer p { margin: 4px 0; }
.footer-tiny { font-size: 12px; opacity: 0.7; }
.hint { font-size: 12px; }
.muted { color: #86909C; }

@media (max-width: 900px) {
  .stage { gap: 30px; }
  .phone { width: 320px; height: 650px; }
}
</style>