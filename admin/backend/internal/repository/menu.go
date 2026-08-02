package repository

import (
	"context"
	"errors"
	"sort"

	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/dto"
	"github.com/admin-template/backend/internal/model"
)

// MenuRepository persists menus and projects them as trees / perm-code sets.
type MenuRepository interface {
	Create(ctx context.Context, m *model.Menu) error
	Update(ctx context.Context, m *model.Menu) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Menu, error)
	ListAll(ctx context.Context) ([]model.Menu, error)
	TreeForRoles(ctx context.Context, roleIDs []uint64) ([]dto.MenuTree, error)
	PermCodesForRoles(ctx context.Context, roleIDs []uint64) ([]string, error)
}

type menuRepo struct{ db *gorm.DB }

func NewMenuRepository(db *gorm.DB) MenuRepository { return &menuRepo{db: db} }

func (r *menuRepo) Create(ctx context.Context, m *model.Menu) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *menuRepo) Update(ctx context.Context, m *model.Menu) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *menuRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("parent_id = ?", id).Delete(&model.Menu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Menu{}, id).Error
	})
}

func (r *menuRepo) GetByID(ctx context.Context, id uint64) (*model.Menu, error) {
	var m model.Menu
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepo) ListAll(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.WithContext(ctx).Order("parent_id ASC, sort ASC").Find(&menus).Error
	return menus, err
}

// TreeForRoles returns the menu tree filtered to the supplied roles. Buttons
// (type=3) are excluded — they are surfaced via PermCodesForRoles.
func (r *menuRepo) TreeForRoles(ctx context.Context, roleIDs []uint64) ([]dto.MenuTree, error) {
	if len(roleIDs) == 0 {
		return []dto.MenuTree{}, nil
	}
	var menus []model.Menu
	err := r.db.WithContext(ctx).
		Table("menus").
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Where("role_menus.role_id IN ? AND menus.type <> ? AND menus.status = ? AND menus.visible = ?",
			roleIDs, model.MenuTypeButton, model.RoleStatusActive, true).
		Order("menus.parent_id ASC, menus.sort ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return buildTree(menus), nil
}

// PermCodesForRoles returns the union of perm_codes across the supplied roles.
// Permission codes live on BOTH routable menus (e.g., "user:view" on the
// user-management page) and buttons (e.g., "user:create" on the create-user
// button) — we collect both. Directories (type=1) carry no perm_code so
// the `perm_code <> ”` filter naturally excludes them.
func (r *menuRepo) PermCodesForRoles(ctx context.Context, roleIDs []uint64) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	var codes []string
	err := r.db.WithContext(ctx).
		Table("menus").
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Where("role_menus.role_id IN ? AND menus.perm_code <> '' AND menus.status = ?",
			roleIDs, model.RoleStatusActive).
		Distinct().
		Pluck("menus.perm_code", &codes).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(codes)
	return codes, nil
}

func buildTree(menus []model.Menu) []dto.MenuTree {
	idx := make(map[uint64]*dto.MenuTree, len(menus))
	roots := []*dto.MenuTree{}
	for i := range menus {
		m := menus[i]
		node := &dto.MenuTree{
			ID: m.ID, ParentID: m.ParentID, Type: int8(m.Type),
			Name: m.Name, Title: m.Title, Path: m.Path, Component: m.Component,
			PermCode: m.PermCode, Icon: m.Icon, Sort: m.Sort, Visible: m.Visible,
		}
		idx[m.ID] = node
		if m.ParentID == 0 {
			roots = append(roots, node)
		}
	}
	for i := range menus {
		m := menus[i]
		if m.ParentID == 0 {
			continue
		}
		if parent, ok := idx[m.ParentID]; ok {
			parent.Children = append(parent.Children, *nodeFromTree(idx[m.ID]))
		}
	}
	out := make([]dto.MenuTree, 0, len(roots))
	for _, r := range roots {
		out = append(out, *r)
	}
	return out
}

func nodeFromTree(t *dto.MenuTree) *dto.MenuTree {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}
