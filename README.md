# 校园灵活用工小程序 — 模板

一个面向校园场景的"灵活用工 / 兼职招聘"小程序模板。学生找兼职,雇主发岗位,平台审核。完整端到端:登录 → 浏览 → 报名 → 审核 → 录用(Mock 支付)→ 上岗 → 完成 → 评价。

```
FlexibleWork/
├── admin/         # Go + Vue 3 后端 + 管理后台
│   ├── backend/   # Go 1.25 + Gin + GORM v2 + MySQL + JWT + RBAC
│   └── frontend/  # Vue 3 + Element Plus (管理端,本模板未做业务扩展)
└── uniapp/        # uni-app 微信小程序 (Vue 3)
```

---

## 一、技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25 · Gin · GORM v2 · MySQL 8 · JWT · RBAC |
| 管理端 | Vue 3 · Element Plus · Vite (admin/,本模板未做业务扩展,沿用原模板) |
| 小程序 | uni-app · Vue 3 · Pinia · 微信小程序 |
| 部署 | Docker Compose (mysql + backend + frontend-nginx) |

---

## 二、目录结构

### 后端 `admin/backend/`

```
backend/
├── cmd/server/main.go              # 二进制入口
├── internal/
│   ├── cmd/run.go                  # 装配:配置 + DB + 路由 + 启动
│   ├── config/                     # env + yaml 加载
│   ├── obs/                        # 日志 + MySQL 池
│   ├── model/                      # GORM struct (8 张业务表 + 原 8 张系统表)
│   │   ├── user.go (扩展 wx_openid)
│   │   ├── profile_student.go
│   │   ├── profile_employer.go
│   │   ├── category.go
│   │   ├── job.go
│   │   ├── application.go
│   │   ├── order.go
│   │   ├── review.go
│   │   └── message.go
│   ├── repository/                 # 8 个业务 repo + 原 6 个系统 repo
│   ├── dto/                        # wire-format 请求/响应
│   ├── service/                    # 业务逻辑 + 状态机
│   ├── api/
│   │   ├── router.go               # 所有路由(public + system + biz + admin)
│   │   ├── handlers/               # 9 个业务 handler + 原 4 个系统 handler
│   │   ├── middleware/             # auth + permission + operation log
│   │   └── httperr/                # 统一错误码
│   ├── pkg/                        # auth / response / storage / pagination / exporter
│   └── seed/seed.go                # 角色 + 菜单 + 示例业务数据
├── configs/config.example.yaml
├── Dockerfile
└── go.mod
```

### 小程序 `uniapp/`

```
uniapp/
├── pages/                          # 17 个页面,4 个 tabbar + 13 个 detail/flow
│   ├── index/                      # 角色感知首页(学生大厅/雇主工作台/未登录)
│   ├── auth/                       # role-select / login / register
│   ├── jobs/                       # 兼职大厅
│   ├── applications/               # 报名(学生)/收到的报名(雇主)
│   ├── orders/                     # 订单(学生/雇主共用,角色感知)
│   ├── mine/                       # 我的
│   ├── job-detail/                 # 岗位详情
│   ├── application-detail/         # 报名详情 + 审核/录用
│   ├── order-detail/               # 订单详情 + 状态机操作
│   ├── certification/              # 认证中心(资料 + 认证)
│   ├── job-publish/                # 发布岗位(雇主)
│   ├── my-jobs/                    # 我的岗位(雇主)
│   ├── dashboard/                  # 雇主工作台
│   ├── messages/                   # 消息中心
│   └── webview/                    # 通用 webview
├── components/
│   ├── job-card/                   # 岗位卡片
│   └── empty-state/                # 空状态
├── api/                            # 接口封装 (auth/category/job/order/profile/message/upload)
├── store/                          # Pinia: user (登录态) + app (分类缓存)
├── utils/                          # constants / auth / format / router / ui
├── static/tabbar/                  # tabbar 图标(占位,需替换为正式设计稿)
├── manifest.json                   # 主题色 + appid
├── pages.json                      # 页面 + tabbar
├── uni.scss                        # 全局 SCSS 变量(品牌色 + 间距 + 字号)
├── App.vue                         # 启动 + 全局样式
└── main.js                         # Vue 3 + Pinia 装配
```

---

## 三、快速开始

### 1. 启动后端

```bash
cd admin
docker compose up -d
```

- MySQL: `localhost:3306` (root/rootpass)
- 后端 API: `http://localhost:8080/api/v1`
- 后端 Swagger: `http://localhost:8080/swagger/index.html`
- 管理端: `http://localhost:8081` (沿用原 admin 模板)

首次启动会自动跑 migration + seed,创建:
- 5 个角色: `super_admin`, `common`, `student`, `employer`, `agent`
- 3 个预置体验账号:
  - 管理员: `admin` / `admin123`
  - 学生: `demo_student` / `demo123` (已通过认证)
  - 雇主: `demo_employer` / `demo123` (已通过资质认证)
- 校园代理 `demo_agent` / `demo123` 是首次注册后由 E2E 测试创建(已通过资质,可以测发岗)
- 7 个分类 + 4 个示例岗位(全部已审核过,status=招聘中)

### 2. 启动小程序

1. 用 HBuilderX 打开 `uniapp/` 目录
2. 修改 `utils/constants.js` 里的 `API_BASE_URL`:
   - Android 模拟器: `http://10.0.2.2:8080`
   - iOS 模拟器: `http://localhost:8080`
   - 真机/小程序开发工具: `http://<你的电脑局域网IP>:8080`
3. 在 `manifest.json` 填入你的微信小程序 appid(`mp-weixin.appid`)
4. 运行 → 运行到小程序模拟器(或真机/微信开发者工具)
5. 用 `demo_student` / `demo_employer` / `demo_agent` 登录体验

---

## 四、业务流转(状态机)

### 岗位 (job)
```
0 草稿 → 1 待审核 → 2 招聘中 → 3 已下架
                    2 → 4 审核未通过(管理端)
                    2 → 5 已招满(自动,录用数 == 招募数)
```

### 报名 (application)
```
1 待审核 → 2 已通过 (雇主) → 5 已转订单 (雇主录用)
          → 3 已拒绝 (雇主)
          → 4 已取消 (学生)
```

### 订单 (order)
```
1 待支付 → 2 已支付(mock) → 3 进行中(学生checkin) → 4 待确认完成(学生complete) → 5 已结算(雇主confirm)
                                                                            ↘ 取消(任一方)→ 6 已取消 / 7 已退款
```

---

## 五、关键接口

### 公共
- `POST /api/v1/auth/register` (user_type: student|employer)
- `POST /api/v1/auth/login` / `POST /api/v1/auth/refresh`
- `GET  /api/v1/categories`
- `GET  /api/v1/jobs?page&size&category_id&location&salary_min&salary_max&keyword`
- `GET  /api/v1/jobs/:id`

### 学生 (需要 `student` 角色 + 对应权限码)
- `GET  /api/v1/student/profile`
- `POST /api/v1/student/profile`
- `POST /api/v1/student/certification`
- `POST /api/v1/jobs/:id/apply`
- `GET  /api/v1/applications` / `:id` / `:id/cancel`
- `GET  /api/v1/orders` / `:id` / `:id/pay` / `:id/checkin` / `:id/complete` / `:id/cancel` / `:id/review`

### 校园代理 (需要 `agent` 角色 + 对应权限码)
- `GET  /api/v1/agent/profile`
- `POST /api/v1/agent/profile`
- `POST /api/v1/agent/certification`

> 校园代理与雇主共享业务权限与端点。发岗、报名审核、下单、订单等都走 `/api/v1/employer/*`(JobService.Create 通过 `resolvePoster()` 同时支持两者),前端根据 `user_type` 自动切换文案。

### 雇主 (需要 `employer` 角色 + 对应权限码)
- `GET  /api/v1/employer/profile` / `POST /employer/profile` / `POST /employer/certification`
- `GET  /api/v1/employer/jobs` / `POST` / `PUT /:id` / `DELETE /:id` / `POST /:id/offline`
- `GET  /api/v1/employer/jobs/:id/applications`
- `POST /api/v1/employer/applications/:id/audit` (action: 2=通过 3=拒绝)
- `POST /api/v1/employer/applications/:id/hire` (创建订单,带金额)
- `GET  /api/v1/employer/orders` / `:id` / `:id/confirm` / `:id/cancel` / `:id/review`

### 管理 (需要 `super_admin` 角色 + 对应权限码)
- `GET/POST/PUT/DELETE /api/v1/admin/categories`
- `GET /api/v1/admin/jobs?status=1` (待审核)
- `POST /api/v1/admin/jobs/:id/audit` (action: 2=通过 4=拒绝)
- `GET/POST /api/v1/admin/student-certifications`
- `GET/POST /api/v1/admin/employer-certifications`
- `GET/POST /api/v1/admin/agent-certifications` (校园代理资质审核)
- `GET /api/v1/admin/orders`
- `GET/DELETE /api/v1/admin/reviews/:id`
- `POST /api/v1/admin/messages/broadcast`

---

## 六、角色与权限码

权限码通过 JWT 签发时烤进 claims,`RequirePerm` 中间件 O(1) 校验。

| 角色 | 权限 |
|---|---|
| `super_admin` | 全部(通过 menu→role 联动 + 全部业务权限) |
| `student` | profile:*, cert:submit, job:apply, application:*, order:*, review:create, message:view |
| `employer` | profile:*, cert:submit, job:create/update/delete/offline, application:*, order:*, review:create, message:view |
| `agent`   | 同 employer(共享业务权限;JobService 通过 `resolvePoster()` 同时支持) |

学生/雇主/代理角色的权限码在 `service/auth_service.go::roleCodePermissions` 里硬编码(为简化模板,不走 menu 系统)。

**校园代理 vs 雇主**:两者都"发岗招人",业务权限完全相同。区别在:
- profile 表不同(`agent_profiles` 多了 `referral_code` 推荐码、`bank_account` 银行预留、`total_referrals/total_earnings` 佣金预留)
- 注册时前端选的角色不同,后端会创建对应的 profile 行
- admin 后台的"资质审核"页有第三个 tab "校园代理资质"
- 小程序首页/我的页:agent 用黄色主题(`$brand-accent`),雇主用绿色主题(`$brand-secondary`),学生用橙色主题(`$brand-primary`)
- "我的"页 agent 多显示一个"我的推荐码"卡片(可一键复制)

---

## 七、扩展点

模板刻意把以下几处留作扩展锚点:

1. **真实支付**: `service/order_service.go::Pay` 当前是 mock。接入微信支付时替换 `payMethod` 字段 + 走 `wechat pay v3` 即可。
2. **WeChat 登录**: `model.User.WxOpenid` 已预留。实现 `POST /auth/wx-login` 时按 `wx_code` 换 `openid` 然后查 user。
3. **管理端业务页面**: 后端 admin 端点已通(`/admin/categories`, `/admin/jobs` 等),Vue 端 `src/views/business/` 目录需要新建(沿用现有 CRUD 模板)。
4. **LBS 附近岗位**: `JobListFilter` 已留位,前端在 `parseJobFilter` 追加 `lat/lng/距离`。
5. **订单状态机的退款**: 当前 6→7 是直接转,真实场景需要接支付平台退款 API。
6. **消息推送**: 当前是 in-app。要接微信订阅消息,加 `template_id` 配置 + 在状态变化时调 `wx.requestSubscribeMessage`。

---

## 八、常见问题

**Q: HBuilderX 报错 "wxss 编译错误"**
A: 升级 HBuilderX 到最新稳定版(3.8+),Vue3 项目需要新版编译器。

**Q: 小程序请求后端报 "不在以下 request 合法域名列表中"**
A: 在微信开发者工具勾选 "不校验合法域名" 用于开发;上线前在 `manifest.json` 配 `mp-weixin.requiredPrivateInfos` 和 `httpsRequest` 合法域名。

**Q: 学生登录后看不到学生端**
A: 登录后查看 Pinia `user.activeRole` 字段;`/auth/me` 返回的 `user_type` 决定走哪个 tabbar。

**Q: 后端改了代码,小程序看不到**
A: docker compose restart backend,小程序重新登录拿新 token。

**Q: 如何重置数据库**
A: `docker compose down -v && docker compose up -d`,会重新跑 migration + seed。

---

## 九、模板展示页 (Showcase)

`http://localhost:8081/showcase` 是个**公开访问、无需登录**的展示页 — 中间是 iPhone 风格的"手机框",里面跑的就是真实的 uni-app 小程序。右侧是特性卡片 + 关键 API 列表,适合直接发链接给 PM / 客户 / 投资人预览。

**当前已实现**(`admin/frontend/src/views/showcase/index.vue`):
- 🎨 渐变背景 + 手机框 mockup(375×760,带刘海)
- 📱 iframe 嵌入 uni-app H5 build(若已构建)
- 🔄 智能 fallback:H5 不存在时自动显示带 demo 账号 + 构建说明的占位页(`hasApp.value = false`)
- 📋 右侧 4 张特性卡片(三端认证 / 状态机 / 管理端 / Docker)
- 🔌 8 条关键 API 速查

### 把真实 H5 build 放进去(可选,但推荐)

展示页默认显示占位。要让 iframe 加载真实的小程序,需要先用 HBuilderX 出 H5:

```bash
# 在 HBuilderX 里:
# 1. 打开 uniapp/ 目录
# 2. 菜单 → 发行 → 网站-PC Web 或 手机H5
# 3. 等待构建完成
# 4. 找到输出目录(unpackage/dist/build/h5/ 或 build/h5/)
# 5. 把里面所有文件复制到 admin/frontend/public/showcase-app/

# 然后重建 admin frontend 容器:
docker compose up -d --build frontend
```

完成后访问 `http://localhost:8081/showcase`,中间的手机框会自动变成真正的小程序 — 登录用 `demo_student` / `demo123` 就能看学生端,切到 `demo_employer` 看雇主端,`demo_agent` 看代理端。

### 配置

| 路径 | 作用 |
|---|---|
| `http://localhost:8081/showcase` | 公开访问的展示页(Vue SPA) |
| `http://localhost:8081/showcase-app/` | uni-app H5 构建产物(via nginx alias) |
| `http://localhost:8081/` | 默认重定向到 `/showcase`(未登录态) |

nginx 配置在 `admin/frontend/nginx.conf` 的 `location /showcase-app/` 块,模板里已写好。docker build 时 `public/` 目录会被 vite 自动复制到 `dist/` 再拷进容器。

**降级行为**:`public/showcase-app/index.html` 没替换时,Vue 页 `hasApp.value = false`,展示占位(含体验账号和构建步骤);有真实 H5 时自动切换到 iframe 加载。两种状态都有完整视觉,不会出 404。

---

## License

GNU GPLv3
