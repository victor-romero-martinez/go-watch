package runner

import (
	"context"
	"testing"

	"gow/internal/testutil"
)

func TestMockCommander(t *testing.T) {
	mock := &testutil.MockCommander{}
	ctx := context.Background()
	cmd := "go"
	args := []string{"run", "main.go"}

	err := mock.Run(ctx, cmd, args)

	if err != nil {
		t.Error("No se esperaba error en el mock")
	}
	if mock.CallCount < 1 {
		t.Error("El comando debería haber sido llamado")
	}
	if mock.LastCommand != "go" {
		t.Errorf("Se esperaba 'go', se obtuvo: %s", mock.LastCommand)
	}
}
