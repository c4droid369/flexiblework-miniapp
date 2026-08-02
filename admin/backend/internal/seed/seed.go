// Package seed populates the database with the minimum dataset needed for the
// app to be usable immediately after first startup: an admin user, the
// default role set, the menu tree, and a small set of business fixtures
// (categories + sample jobs + sample student/employer accounts) so the
// mini-program has something to render out-of-the-box.
package seed

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// Options controls seed-time credentials.
type Options struct {
	AdminUsername string
	AdminPassword string
}

// Run is idempotent: it inserts missing rows and skips existing ones, so it is
// safe to call on every startup.
func Run(ctx context.Context, db *gorm.DB, logger *slog.Logger, opt Options) error {
	if db == nil {
		return errors.New("nil db")
	}
	if opt.AdminUsername == "" || opt.AdminPassword == "" {
		return errors.New("seed: missing admin credentials")
	}

	if err := seedRoles(ctx, db); err != nil {
		return err
	}
	if err := seedMenus(ctx, db); err != nil {
		return err
	}
	if err := seedAdmin(ctx, db, opt); err != nil {
		return err
	}
	if err := seedBusinessFixtures(ctx, db); err != nil {
		return err
	}
	logger.InfoContext(
		ctx, "seed completed",
		slog.String("admin", opt.AdminUsername),
		slog.String("warning", "CHANGE THE DEFAULT ADMIN PASSWORD BEFORE PRODUCTION"),
	)
	return nil
}

func seedRoles(ctx context.Context, db *gorm.DB) error {
	roles := []model.Role{
		{Code: "super_admin", Name: "超级管理员", Description: "拥有全部权限", Sort: 1, Status: model.RoleStatusActive},
		{Code: "common", Name: "普通用户", Description: "仅可访问仪表盘", Sort: 99, Status: model.RoleStatusActive},
		// Business roles — permissions are hardcoded in AuthService per role
		// code, so the menu binding is not used for these.
		{Code: "student", Name: "学生", Description: "学生端(报名兼职、查看订单、评价)", Sort: 10, Status: model.RoleStatusActive},
		{Code: "employer", Name: "雇主", Description: "雇主端(发布岗位、审核报名、结算订单)", Sort: 11, Status: model.RoleStatusActive},
		{Code: "agent", Name: "校园代理", Description: "校园代理(发代理岗、推广追踪、佣金结算预留)", Sort: 12, Status: model.RoleStatusActive},
	}
	for i := range roles {
		r := roles[i]
		err := db.WithContext(ctx).
			Where("code = ?", r.Code).
			Attrs(r).
			FirstOrCreate(&r).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func seedMenus(ctx context.Context, db *gorm.DB) error {
	// Tree shape:
	//   Dashboard (menu)
	//   System   (directory)
	//     ├ User / Role / Menu / Log  (admin RBAC)
	//   Business (directory) — added for the gig work mini-program
	//     ├ Category (menu + buttons)
	//     ├ Job Audit (menu + audit button)
	//     ├ Cert Audit (menu)
	//     ├ Order (menu, view-only)
	//     ├ Review (menu + delete)
	//     └ Message (menu + broadcast)
	specs := menuSpec()
	for _, s := range specs {
		if err := upsertMenu(ctx, db, s, 0); err != nil {
			return err
		}
	}
	return bindSuperAdminMenus(ctx, db)
}

func upsertMenu(ctx context.Context, db *gorm.DB, s menuSpecT, parentID uint64) error {
	m := model.Menu{
		ParentID:  parentID,
		Type:      s.Type,
		Name:      s.Name,
		Title:     s.Title,
		Path:      s.Path,
		Component: s.Component,
		PermCode:  s.PermCode,
		Icon:      s.Icon,
		Sort:      s.Sort,
		Visible:   s.Visible,
		Status:    model.RoleStatusActive,
	}
	err := db.WithContext(ctx).
		Where("name = ? AND parent_id = ?", m.Name, parentID).
		Assign(m).
		FirstOrCreate(&m).Error
	if err != nil {
		return err
	}
	for _, child := range s.Children {
		if err := upsertMenu(ctx, db, child, m.ID); err != nil {
			return err
		}
	}
	return nil
}

func bindSuperAdminMenus(ctx context.Context, db *gorm.DB) error {
	var super model.Role
	if err := db.WithContext(ctx).Where("code = ?", "super_admin").First(&super).Error; err != nil {
		return err
	}
	var menus []model.Menu
	if err := db.WithContext(ctx).Find(&menus).Error; err != nil {
		return err
	}
	for _, m := range menus {
		rm := model.RoleMenu{RoleID: super.ID, MenuID: m.ID}
		if err := db.WithContext(ctx).
			Where("role_id = ? AND menu_id = ?", rm.RoleID, rm.MenuID).
			Attrs(rm).
			FirstOrCreate(&rm).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAdmin(ctx context.Context, db *gorm.DB, opt Options) error {
	var super model.Role
	if err := db.WithContext(ctx).Where("code = ?", "super_admin").First(&super).Error; err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(opt.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var u model.User
	err = db.WithContext(ctx).
		Where("username = ?", opt.AdminUsername).
		Attrs(model.User{
			Username:     opt.AdminUsername,
			PasswordHash: string(hash),
			Nickname:     "Administrator",
			Email:        "admin@example.com",
			Status:       model.UserStatusActive,
		}).
		FirstOrCreate(&u).Error
	if err != nil {
		return err
	}
	ur := model.UserRole{UserID: u.ID, RoleID: super.ID}
	return db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", ur.UserID, ur.RoleID).
		Attrs(ur).
		FirstOrCreate(&ur).Error
}

// seedBusinessFixtures inserts a handful of categories, one demo employer
// (cert_status=已通过), one demo student, and a few published jobs so the
// mini-program has data to render on first launch. Idempotent.
func seedBusinessFixtures(ctx context.Context, db *gorm.DB) error {
	if err := seedCategories(ctx, db); err != nil {
		return err
	}
	if err := seedDemoAccounts(ctx, db); err != nil {
		return err
	}
	return seedDemoJobs(ctx, db)
}

func seedCategories(ctx context.Context, db *gorm.DB) error {
	cats := []model.Category{
		{Name: "促销派单", Icon: "icon-promote", Sort: 1, Status: 1, Description: "商场、展会派发传单与样品"},
		{Name: "家教", Icon: "icon-tutor", Sort: 2, Status: 1, Description: "中小学一对一/小班辅导"},
		{Name: "餐饮服务", Icon: "icon-restaurant", Sort: 3, Status: 1, Description: "餐厅、咖啡店、奶茶店服务生"},
		{Name: "活动执行", Icon: "icon-event", Sort: 4, Status: 1, Description: "演出、赛事、论坛现场执行"},
		{Name: "线上兼职", Icon: "icon-online", Sort: 5, Status: 1, Description: "文案、翻译、设计、剪辑等远程任务"},
		{Name: "校园代理", Icon: "icon-campus", Sort: 6, Status: 1, Description: "驾校、考研、考证等校园代理"},
		{Name: "其他", Icon: "icon-else", Sort: 99, Status: 1, Description: "其他类型兼职"},
	}
	for i := range cats {
		c := cats[i]
		err := db.WithContext(ctx).Where("name = ?", c.Name).
			Attrs(c).FirstOrCreate(&c).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func seedDemoAccounts(ctx context.Context, db *gorm.DB) error {
	// Demo employer — username "demo_employer" / password "demo123".
	empRole, err := firstRoleByCode(ctx, db, "employer")
	if err != nil {
		return err
	}
	empHash, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	var empUser model.User
	err = db.WithContext(ctx).Where("username = ?", "demo_employer").
		Attrs(model.User{
			Username: "demo_employer", PasswordHash: string(empHash),
			Nickname: "校园咖啡小馆", Email: "employer@demo.com",
			Status: model.UserStatusActive,
		}).FirstOrCreate(&empUser).Error
	if err != nil {
		return err
	}
	_ = db.WithContext(ctx).Where("user_id = ? AND role_id = ?", empUser.ID, empRole.ID).
		Attrs(model.UserRole{UserID: empUser.ID, RoleID: empRole.ID}).
		FirstOrCreate(&model.UserRole{}).Error
	empProfile := model.EmployerProfile{
		UserID: empUser.ID, CompanyName: "校园咖啡小馆", ContactName: "李老板",
		ContactPhone: "13800000001", ContactEmail: "boss@demo.com",
		BusinessLicenseNo: "BL-2026-DEMO-0001", BusinessLicenseImg: "/files/demo/bl.png",
		Industry: "餐饮", CompanySize: "1-10", CompanyAddress: "学校东门商业街 12 号",
		Intro:         "面向学生的精品咖啡馆,招聘靠谱学生兼职",
		CertStatus:    2, // 已通过 — demo 直接给通过
		CertifiedAt:   ptrTime(time.Now()),
		Rating:        4.8, TotalJobs: 5, CompletedOrders: 12,
	}
	_ = db.WithContext(ctx).Where("user_id = ?", empUser.ID).
		Attrs(empProfile).FirstOrCreate(&model.EmployerProfile{}).Error

	// Demo student — username "demo_student" / password "demo123".
	stuRole, err := firstRoleByCode(ctx, db, "student")
	if err != nil {
		return err
	}
	stuHash, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	var stuUser model.User
	err = db.WithContext(ctx).Where("username = ?", "demo_student").
		Attrs(model.User{
			Username: "demo_student", PasswordHash: string(stuHash),
			Nickname: "小明同学", Email: "student@demo.com",
			Status: model.UserStatusActive,
		}).FirstOrCreate(&stuUser).Error
	if err != nil {
		return err
	}
	_ = db.WithContext(ctx).Where("user_id = ? AND role_id = ?", stuUser.ID, stuRole.ID).
		Attrs(model.UserRole{UserID: stuUser.ID, RoleID: stuRole.ID}).
		FirstOrCreate(&model.UserRole{}).Error
	stuProfile := model.StudentProfile{
		UserID: stuUser.ID, RealName: "王小明", Gender: 1,
		School: "示例大学", College: "计算机学院", Major: "软件工程",
		StudentNo: "2023001001", IDCardNo: "110000202301010001",
		IDCardFront: "/files/demo/id_front.png", IDCardBack: "/files/demo/id_back.png",
		StudentCard: "/files/demo/student_card.png",
		CertStatus: 2, CertifiedAt: ptrTime(time.Now()),
		Bio:    "踏实靠谱,周末可全天",
		Skills: `["认真负责","沟通能力好"]`,
	}
	_ = db.WithContext(ctx).Where("user_id = ?", stuUser.ID).
		Attrs(stuProfile).FirstOrCreate(&model.StudentProfile{}).Error

	return nil
}

func seedDemoJobs(ctx context.Context, db *gorm.DB) error {
	// Find the demo employer + the first 4 categories.
	var empUser model.User
	if err := db.WithContext(ctx).Where("username = ?", "demo_employer").First(&empUser).Error; err != nil {
		return err // no demo employer, skip
	}
	var cats []model.Category
	if err := db.WithContext(ctx).Order("sort ASC").Limit(4).Find(&cats).Error; err != nil {
		return err
	}
	if len(cats) == 0 {
		return nil
	}
	jobs := []model.Job{
		{
			EmployerID: empUser.ID, CategoryID: cats[0].ID,
			Title:        "周末饮品店店员",
			Description:  "负责点单、出品、清洁。需站立工作 4-6 小时。提供员工餐。",
			Requirements: "性格开朗,能连续站立,周末可到岗",
			SalaryType:   2, SalaryMin: 150, SalaryMax: 200, SalaryUnit: "元/天",
			Location:     "学校东门商业街 12 号", WorkDateType: 2,
			RecruitCount: 2, GenderRequirement: 0, SettlementType: 1,
			Status: 2, ViewCount: 123, ApplyCount: 5,
		},
		{
			EmployerID: empUser.ID, CategoryID: cats[1%len(cats)].ID,
			Title:        "高一数学周末家教",
			Description:  "为高一学生补习数学,讲解学校进度内容。每次 2 小时,持续一学期。",
			Requirements: "数学相关专业,有耐心",
			SalaryType:   1, SalaryMin: 80, SalaryMax: 100, SalaryUnit: "元/小时",
			Location:     "学生宿舍区", WorkDateType: 3,
			RecruitCount: 1, GenderRequirement: 0, SettlementType: 3,
			Status: 2, ViewCount: 87, ApplyCount: 3,
		},
		{
			EmployerID: empUser.ID, CategoryID: cats[2%len(cats)].ID,
			Title:        "校园活动现场协助",
			Description:  "新生晚会现场指引、签到、物资搬运。",
			Requirements: "男生优先,能搬 10kg 左右物资",
			SalaryType:   2, SalaryMin: 120, SalaryMax: 150, SalaryUnit: "元/天",
			Location:     "学校体育馆", WorkDateType: 3,
			WorkTimeStart: "18:00", WorkTimeEnd: "22:00",
			RecruitCount: 5, GenderRequirement: 1, SettlementType: 1,
			Status: 2, ViewCount: 56, ApplyCount: 2,
		},
		{
			EmployerID: empUser.ID, CategoryID: cats[3%len(cats)].ID,
			Title:        "公众号推文排版与配图",
			Description:  "学校官方公众号日常推文排版,寻找合适配图。要求熟悉秀米。",
			Requirements: "每周 2-3 篇,可远程",
			SalaryType:   4, SalaryMin: 800, SalaryMax: 1200, SalaryUnit: "元/月",
			Location:     "远程", WorkDateType: 1,
			RecruitCount: 1, GenderRequirement: 0, SettlementType: 3,
			Status: 2, ViewCount: 42, ApplyCount: 1,
		},
	}
	for i := range jobs {
		j := jobs[i]
		err := db.WithContext(ctx).Where("title = ? AND employer_id = ?", j.Title, j.EmployerID).
			Attrs(j).FirstOrCreate(&model.Job{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func firstRoleByCode(ctx context.Context, db *gorm.DB, code string) (model.Role, error) {
	var r model.Role
	if err := db.WithContext(ctx).Where("code = ?", code).First(&r).Error; err != nil {
		return r, err
	}
	return r, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

type menuSpecT struct {
	Type      model.MenuType
	Name      string
	Title     string
	Path      string
	Component string
	PermCode  string
	Icon      string
	Sort      int
	Visible   bool
	Children  []menuSpecT
}

func menuSpec() []menuSpecT {
	return []menuSpecT{
		{Type: model.MenuTypeMenu, Name: "Dashboard", Title: "仪表盘", Path: "/dashboard", Component: "views/dashboard/index.vue", Icon: "dashboard", Sort: 1, Visible: true},
		{Type: model.MenuTypeDirectory, Name: "System", Title: "系统管理", Path: "/system", Icon: "setting", Sort: 99, Visible: true, Children: []menuSpecT{
			{Type: model.MenuTypeMenu, Name: "User", Title: "用户管理", Path: "/system/user", Component: "views/system/user/index.vue", PermCode: "user:view", Icon: "user", Sort: 1, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "UserCreate", Title: "创建用户", PermCode: "user:create", Sort: 1},
				{Type: model.MenuTypeButton, Name: "UserUpdate", Title: "编辑用户", PermCode: "user:update", Sort: 2},
				{Type: model.MenuTypeButton, Name: "UserDelete", Title: "删除用户", PermCode: "user:delete", Sort: 3},
				{Type: model.MenuTypeButton, Name: "UserResetPassword", Title: "重置密码", PermCode: "user:reset-password", Sort: 4},
			}},
			{Type: model.MenuTypeMenu, Name: "Role", Title: "角色管理", Path: "/system/role", Component: "views/system/role/index.vue", PermCode: "role:view", Icon: "userFilled", Sort: 2, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "RoleCreate", Title: "创建角色", PermCode: "role:create", Sort: 1},
				{Type: model.MenuTypeButton, Name: "RoleUpdate", Title: "编辑角色", PermCode: "role:update", Sort: 2},
				{Type: model.MenuTypeButton, Name: "RoleDelete", Title: "删除角色", PermCode: "role:delete", Sort: 3},
				{Type: model.MenuTypeButton, Name: "RoleAssignMenu", Title: "分配菜单", PermCode: "role:assign-menu", Sort: 4},
				{Type: model.MenuTypeButton, Name: "RoleAssignUser", Title: "分配用户", PermCode: "role:assign-user", Sort: 5},
			}},
			{Type: model.MenuTypeMenu, Name: "Menu", Title: "菜单管理", Path: "/system/menu", Component: "views/system/menu/index.vue", PermCode: "menu:view", Icon: "menu", Sort: 3, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "MenuCreate", Title: "创建菜单", PermCode: "menu:create", Sort: 1},
				{Type: model.MenuTypeButton, Name: "MenuUpdate", Title: "编辑菜单", PermCode: "menu:update", Sort: 2},
				{Type: model.MenuTypeButton, Name: "MenuDelete", Title: "删除菜单", PermCode: "menu:delete", Sort: 3},
			}},
			{Type: model.MenuTypeMenu, Name: "Log", Title: "操作日志", Path: "/system/log", Component: "views/system/log/index.vue", PermCode: "log:view", Icon: "document", Sort: 4, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "LogDelete", Title: "批量删除", PermCode: "log:delete", Sort: 1},
			}},
		}},
		// Business — campus gig work management (admin side).
		{Type: model.MenuTypeDirectory, Name: "Business", Title: "业务管理", Path: "/business", Icon: "briefcase", Sort: 50, Visible: true, Children: []menuSpecT{
			{Type: model.MenuTypeMenu, Name: "Category", Title: "兼职分类", Path: "/business/category", Component: "views/business/category/index.vue", PermCode: "category:view", Icon: "menu", Sort: 1, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "CategoryCreate", Title: "创建分类", PermCode: "category:create", Sort: 1},
				{Type: model.MenuTypeButton, Name: "CategoryUpdate", Title: "编辑分类", PermCode: "category:update", Sort: 2},
				{Type: model.MenuTypeButton, Name: "CategoryDelete", Title: "删除分类", PermCode: "category:delete", Sort: 3},
			}},
			{Type: model.MenuTypeMenu, Name: "JobAudit", Title: "岗位审核", Path: "/business/job-audit", Component: "views/business/job-audit/index.vue", PermCode: "job:audit", Icon: "document-checked", Sort: 2, Visible: true},
			{Type: model.MenuTypeMenu, Name: "CertAudit", Title: "资质审核", Path: "/business/cert-audit", Component: "views/business/cert-audit/index.vue", PermCode: "cert:audit", Icon: "user-checked", Sort: 3, Visible: true},
			{Type: model.MenuTypeMenu, Name: "OrderMonitor", Title: "订单监控", Path: "/business/order", Component: "views/business/order/index.vue", PermCode: "order:monitor", Icon: "list", Sort: 4, Visible: true},
			{Type: model.MenuTypeMenu, Name: "ReviewManage", Title: "评价管理", Path: "/business/review", Component: "views/business/review/index.vue", PermCode: "review:view", Icon: "chat", Sort: 5, Visible: true, Children: []menuSpecT{
				{Type: model.MenuTypeButton, Name: "ReviewDelete", Title: "删除评价", PermCode: "review:delete", Sort: 1},
			}},
			{Type: model.MenuTypeMenu, Name: "MessageManage", Title: "消息管理", Path: "/business/message", Component: "views/business/message/index.vue", PermCode: "message:broadcast", Icon: "bell", Sort: 6, Visible: true},
		}},
		// File management — kept off the sidebar (component="") but its
		// permission codes still reach super_admin's role_menus through the
		// menu tree.
		{Type: model.MenuTypeButton, Name: "FileUpload", Title: "上传文件", PermCode: "file:upload", Sort: 1},
		{Type: model.MenuTypeButton, Name: "FileDelete", Title: "删除文件", PermCode: "file:delete", Sort: 2},
	}
}
