package validation

import "testing"

func TestValidateStaffInput(t *testing.T) {
	input, err := ValidateStaffInput(StaffInput{Name: "  林  青 ", Phone: "13800138000", Email: "LIN@EXAMPLE.COM", Role: "photographer"})
	if err != nil {
		t.Fatal(err)
	}
	if input.Name != "林 青" || input.Email != "lin@example.com" {
		t.Fatalf("unexpected normalization: %+v", input)
	}
	_, err = ValidateStaffInput(StaffInput{Name: "", Role: "owner"})
	if err == nil {
		t.Fatal("expected validation failure")
	}
}
