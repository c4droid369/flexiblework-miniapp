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
- 4 个角色: `super_admin`, `common`, `student`, `employer`
- 3 个体验账号:
  - 管理员: `admin` / `admin123`
  - 学生: `demo_student` / `demo123` (已通过认证)
  - 雇主: `demo_employer` / `demo123` (已通过资质认证)
- 7 个分类 + 4 个示例岗位(全部已审核过,status=招聘中)

### 2. 启动小程序

1. 用 HBuilderX 打开 `uniapp/` 目录
2. 修改 `utils/constants.js` 里的 `API_BASE_URL`:
   - Android 模拟器: `http://10.0.2.2:8080`
   - iOS 模拟器: `http://localhost:8080`
   - 真机/小程序开发工具: `http://<你的电脑局域网IP>:8080`
3. 在 `manifest.json` 填入你的微信小程序 appid(`mp-weixin.appid`)
4. 运行 → 运行到小程序模拟器(或真机/微信开发者工具)
5. 用 `demo_student` 或 `demo_employer` 登录体验

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

学生/雇主角色的权限码在 `service/auth_service.go::roleCodePermissions` 里硬编码(为简化模板,不走 menu 系统)。

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

## License

GNU GPLv3
