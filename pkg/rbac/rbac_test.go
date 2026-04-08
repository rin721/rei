package rbac

import "testing"

func TestManagerAssignRoleAndCheckPermission(t *testing.T) {
	t.Parallel()

	manager, err := New(Config{})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := manager.AddPolicy("admin", "/reports", "read"); err != nil {
		t.Fatalf("AddPolicy() returned error: %v", err)
	}
	if err := manager.AssignRole("alice", "admin"); err != nil {
		t.Fatalf("AssignRole() returned error: %v", err)
	}

	allowed, err := manager.CheckPermission("alice", "/reports", "read")
	if err != nil {
		t.Fatalf("CheckPermission() returned error: %v", err)
	}
	if !allowed {
		t.Fatal("CheckPermission() = false, want true")
	}

	roles, err := manager.GetUserRoles("alice")
	if err != nil {
		t.Fatalf("GetUserRoles() returned error: %v", err)
	}
	if len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("roles = %#v, want %#v", roles, []string{"admin"})
	}

	users, err := manager.GetUsersForRole("admin")
	if err != nil {
		t.Fatalf("GetUsersForRole() returned error: %v", err)
	}
	if len(users) != 1 || users[0] != "alice" {
		t.Fatalf("users = %#v, want %#v", users, []string{"alice"})
	}
}
