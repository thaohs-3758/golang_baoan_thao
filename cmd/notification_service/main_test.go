package main

import (
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

func TestNotificationServiceBootstrap_BuildsDependencies(t *testing.T) {
	repo := repositories.NewNotificationRepository(nil)
	if repo == nil {
		t.Fatal("expected notification repository instance")
	}
}
