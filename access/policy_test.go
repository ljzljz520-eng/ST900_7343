package access

import "testing"

func TestDirectoryAuthorization(t *testing.T) {
	viewer, err := NewPrincipal("viewer-1", AdminViewer)
	if err != nil {
		t.Fatal(err)
	}
	directory, err := NewDirectory([]Principal{viewer})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := directory.Authorize("viewer-1", PermissionStaffRead); err != nil {
		t.Fatal(err)
	}
	if _, err := directory.Authorize("viewer-1", PermissionStaffDisable); err == nil {
		t.Fatal("expected permission denial")
	}
}
